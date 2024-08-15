package SystemgeServer

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type AsyncMessageHandler func(*Message.Message) error
type SyncMessageHandler func(*Message.Message) (string, error)

type SystemgeServer struct {
	status      int
	statusMutex sync.Mutex

	config *Config.SystemgeServer

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer
	ipRateLimiter *Tools.IpRateLimiter

	tcpServer *Tcp.Listener

	asyncMessageHandlers map[string]AsyncMessageHandler
	syncMessageHandlers  map[string]SyncMessageHandler
	clientConnections    map[string]*clientConnection // name -> clientConnection

	stopChannel                            chan bool //closing of this channel initiates the stop of the systemge component
	clientConnectionStopChannelStopChannel chan bool //closing of this channel indicates that the client connection handler has stopped
	allClientConnectionsStoppedChannel     chan bool //closing of this channel indicates that all client connections have stopped
	messageHandlerChannel                  chan func()

	clientConnectionMutex    sync.RWMutex
	syncMessageHandlerMutex  sync.RWMutex
	asyncMessageHandlerMutex sync.RWMutex

	// metrics
	bytesReceived           atomic.Uint64
	bytesSent               atomic.Uint64
	invalidMessagesReceived atomic.Uint32

	messageRateLimiterExceeded atomic.Uint32
	byteRateLimiterExceeded    atomic.Uint32

	connectionAttempts             atomic.Uint32
	connectionAttemptsSuccessful   atomic.Uint32
	connectionAttemptsFailed       atomic.Uint32
	connectionAttemptBytesSent     atomic.Uint64
	connectionAttemptBytesReceived atomic.Uint64

	syncSuccessResponsesSent atomic.Uint32
	syncFailureResponsesSent atomic.Uint32
	syncResponseBytesSent    atomic.Uint64

	asyncMessagesReceived     atomic.Uint32
	asyncMessageBytesReceived atomic.Uint64

	syncRequestsReceived     atomic.Uint32
	syncRequestBytesReceived atomic.Uint64

	topicAddSent    atomic.Uint32
	topicRemoveSent atomic.Uint32
}

func New(config *Config.SystemgeServer, asyncMessageHanlders map[string]AsyncMessageHandler, syncMessageHandlers map[string]SyncMessageHandler) *SystemgeServer {
	return &SystemgeServer{
		config:               config,
		infoLogger:           Tools.NewLogger("[Info: \""+config.Name+"\"] ", config.InfoLoggerPath),
		warningLogger:        Tools.NewLogger("[Warning: \""+config.Name+"\"] ", config.WarningLoggerPath),
		errorLogger:          Tools.NewLogger("[Error: \""+config.Name+"\"] ", config.ErrorLoggerPath),
		mailer:               Tools.NewMailer(config.MailerConfig),
		clientConnections:    make(map[string]*clientConnection),
		asyncMessageHandlers: asyncMessageHanlders,
		syncMessageHandlers:  syncMessageHandlers,
	}
}

func (server *SystemgeServer) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STOPPED {
		return Error.New("SystemgeServer is not in stopped state", nil)
	}
	server.status = Status.PENDING

	if server.config.TcpBufferBytes == 0 {
		server.config.TcpBufferBytes = 1024 * 4
	}
	tcpServer, err := Tcp.NewServer(server.config.ServerConfig)
	if err != nil {
		server.status = Status.STOPPED
		return Error.New("Failed to create tcp server", err)
	}
	if server.config.IpRateLimiter != nil {
		server.ipRateLimiter = Tools.NewIpRateLimiter(server.config.IpRateLimiter)
	}

	server.tcpServer = tcpServer
	server.stopChannel = make(chan bool)
	server.clientConnectionStopChannelStopChannel = make(chan bool)
	server.allClientConnectionsStoppedChannel = make(chan bool)

	if server.config.ProcessAllMessagesSequentially {
		server.messageHandlerChannel = make(chan func(), server.config.ProcessAllMessagesSequentiallyChannelSize)
		if server.config.ProcessAllMessagesSequentiallyChannelSize == 0 {
			go func() {
				for {
					select {
					case f := <-server.messageHandlerChannel:
						f()
					case <-server.stopChannel:
						return
					}
				}
			}()
		} else {
			go func() {
				for {
					select {
					case f := <-server.messageHandlerChannel:
						if server.errorLogger != nil && len(server.messageHandlerChannel) >= server.config.ProcessAllMessagesSequentiallyChannelSize-1 {
							server.errorLogger.Log("ProcessAllMessagesSequentiallyChannelSize reached (increase ProcessAllMessagesSequentiallyChannelSize otherwise message order of arrival is not guaranteed)")
						}
						f()
					case <-server.allClientConnectionsStoppedChannel:
						return
					}
				}
			}()
		}

	}
	go server.handleClientConnections()
	server.status = Status.STARTED
	return nil
}

// stopSystemgeServerComponent stops the systemge component.
// blocking until all goroutines associated with the systemge component have stopped.
func (server *SystemgeServer) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STARTED {
		return Error.New("SystemgeServer is not in started state", nil)
	}
	server.status = Status.PENDING

	close(server.stopChannel)
	if server.ipRateLimiter != nil {
		server.ipRateLimiter.Stop()
	}

	server.tcpServer.GetListener().Close()
	<-server.clientConnectionStopChannelStopChannel

	server.clientConnectionMutex.Lock()
	for _, clientConnection := range server.clientConnections {
		clientConnection.netConn.Close()
		<-clientConnection.stopChannel
	}
	server.clientConnectionMutex.Unlock()
	close(server.allClientConnectionsStoppedChannel)

	server.status = Status.STOPPED
	return nil
}

func (server *SystemgeServer) GetName() string {
	return server.config.Name
}

func (server *SystemgeServer) GetStatus() int {
	return server.status
}
