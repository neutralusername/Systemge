package TcpSystemgeConnection

import (
	"encoding/json"
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type messageInProcess struct {
	message *Message.Message
	id      uint64
}

type TcpConnection struct {
	name       string
	config     *Config.TcpSystemgeConnection
	netConn    net.Conn
	randomizer *Tools.Randomizer

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer

	sendMutex    sync.Mutex
	receiveMutex sync.Mutex

	closed      bool
	closedMutex sync.Mutex

	tcpBuffer []byte

	syncRequests map[string]*syncRequestStruct
	syncMutex    sync.Mutex

	closeChannel chan bool

	processMutex               sync.Mutex
	processingChannel          chan *messageInProcess
	processingChannelSemaphore *Tools.Semaphore
	processingLoopStopChannel  chan bool

	rateLimiterBytes    *Tools.TokenBucketRateLimiter
	rateLimiterMessages *Tools.TokenBucketRateLimiter

	messageId uint64

	// metrics
	bytesSent     atomic.Uint64
	bytesReceived atomic.Uint64

	asyncMessagesSent atomic.Uint64
	syncRequestsSent  atomic.Uint64

	syncSuccessResponsesReceived atomic.Uint64
	syncFailureResponsesReceived atomic.Uint64
	noSyncResponseReceived       atomic.Uint64
	invalidSyncResponsesReceived atomic.Uint64

	invalidMessagesReceived atomic.Uint64
	validMessagesReceived   atomic.Uint64

	messageRateLimiterExceeded atomic.Uint64
	byteRateLimiterExceeded    atomic.Uint64
}

func New(name string, config *Config.TcpSystemgeConnection, netConn net.Conn) *TcpConnection {
	connection := &TcpConnection{
		name:                       name,
		config:                     config,
		netConn:                    netConn,
		randomizer:                 Tools.NewRandomizer(config.RandomizerSeed),
		closeChannel:               make(chan bool),
		syncRequests:               make(map[string]*syncRequestStruct),
		processingChannel:          make(chan *messageInProcess, config.ProcessingChannelCapacity+1), // +1 so that the receive loop is never blocking while adding a message to the processing channel
		processingChannelSemaphore: Tools.NewSemaphore(config.ProcessingChannelCapacity+1, config.ProcessingChannelCapacity+1),
	}
	if config.InfoLoggerPath != "" {
		connection.infoLogger = Tools.NewLogger("[Info: \""+name+"\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		connection.warningLogger = Tools.NewLogger("[Warning: \""+name+"\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		connection.errorLogger = Tools.NewLogger("[Error: \""+name+"\"] ", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		connection.mailer = Tools.NewMailer(config.MailerConfig)
	}
	if config.RateLimiterBytes != nil {
		connection.rateLimiterBytes = Tools.NewTokenBucketRateLimiter(config.RateLimiterBytes)
	}
	if config.RateLimiterMessages != nil {
		connection.rateLimiterMessages = Tools.NewTokenBucketRateLimiter(config.RateLimiterMessages)
	}
	if config.TcpBufferBytes <= 0 {
		config.TcpBufferBytes = 1024 * 4
	}
	go connection.addMessageToProcessingChannelLoop()
	if config.HeartbeatIntervalMs > 0 {
		go connection.heartbeatLoop()
	}
	return connection
}

func (connection *TcpConnection) Close() error {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()

	if connection.closed {
		return Error.New("Connection already closed", nil)
	}

	connection.closed = true
	close(connection.closeChannel)
	connection.netConn.Close()

	if connection.rateLimiterBytes != nil {
		connection.rateLimiterBytes.Close()
		connection.rateLimiterBytes = nil
	}
	if connection.rateLimiterMessages != nil {
		connection.rateLimiterMessages.Close()
		connection.rateLimiterMessages = nil
	}

	close(connection.processingChannel)
	return nil
}

func (connection *TcpConnection) GetStatus() int {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()
	if connection.closed {
		return Status.STOPPED
	} else {
		return Status.STARTED
	}
}

func (connection *TcpConnection) GetName() string {
	return connection.name
}

// GetCloseChannel returns a channel that will be closed when the connection is closed.
// Blocks until the connection is closed.
// This can be used to trigger an event when the connection is closed.
func (connection *TcpConnection) GetCloseChannel() <-chan bool {
	return connection.closeChannel
}

func (connection *TcpConnection) GetAddress() string {
	return connection.netConn.RemoteAddr().String()
}

func (connection *TcpConnection) GetDefaultCommands() Commands.Handlers {
	commands := Commands.Handlers{}
	commands["close"] = func(args []string) (string, error) {
		err := connection.Close()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["getStatus"] = func(args []string) (string, error) {
		return Status.ToString(connection.GetStatus()), nil
	}
	commands["getMetrics"] = func(args []string) (string, error) {
		metrics := connection.GetMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Error.New("failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	commands["retrieveMetrics"] = func(args []string) (string, error) {
		metrics := connection.RetrieveMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Error.New("failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	commands["unprocessedMessageCount"] = func(args []string) (string, error) {
		return Helpers.Uint32ToString(connection.processingChannelSemaphore.AvailablePermits()), nil
	}
	commands["getNextMessage"] = func(args []string) (string, error) {
		message, err := connection.GetNextMessage()
		if err != nil {
			return "", err
		}
		return string(message.Serialize()), nil
	}
	commands["stopProcessingLoop"] = func(args []string) (string, error) {
		err := connection.StopProcessingLoop()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["asyncMessage"] = func(args []string) (string, error) {
		if len(args) != 2 {
			return "", Error.New("expected at least 2 arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		err := connection.AsyncMessage(topic, payload)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["syncRequest"] = func(args []string) (string, error) {
		if len(args) != 2 {
			return "", Error.New("expected at least 2 arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		message, err := connection.SyncRequestBlocking(topic, payload)
		if err != nil {
			return "", err
		}
		return string(message.Serialize()), nil
	}
	commands["syncResponse"] = func(args []string) (string, error) {
		if len(args) != 3 {
			return "", Error.New("expected 2 arguments", nil)
		}
		message, err := Message.Deserialize([]byte(args[0]), "")
		if err != nil {
			return "", Error.New("failed to deserialize message", err)
		}
		responsePayload := args[1]
		success := args[2] == "true"
		err = connection.SyncResponse(message, success, responsePayload)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["openSyncRequests"] = func(args []string) (string, error) {
		openSyncRequest := connection.GetOpenSyncRequests()
		json, err := json.Marshal(openSyncRequest)
		if err != nil {
			return "", Error.New("failed to marshal open sync requests to json", err)
		}
		return string(json), nil
	}
	commands["abortSyncRequest"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("expected 1 argument", nil)
		}
		err := connection.AbortSyncRequest(args[0])
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	rateLimiterBytesCommands := connection.rateLimiterBytes.GetDefaultCommands()
	for key, value := range rateLimiterBytesCommands {
		commands["rateLimiterBytes_"+key] = value
	}
	rateLimiterMessagesCommands := connection.rateLimiterMessages.GetDefaultCommands()
	for key, value := range rateLimiterMessagesCommands {
		commands["rateLimiterMessages_"+key] = value
	}
	return commands
}
