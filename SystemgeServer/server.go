package Node

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type AsyncMessageHandler func(*Message.Message) error
type SyncMessageHandler func(*Message.Message) (string, error)

type SystemgeServer struct {
	config *Config.SystemgeServer

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer
	ipRateLimiter *Tools.IpRateLimiter

	tcpServer *Tcp.Server

	asyncMessageHandlers map[string]AsyncMessageHandler
	syncMessageHandlers  map[string]SyncMessageHandler
	incomingConnections  map[string]*incomingConnection // name -> nodeConnection

	stopChannel                          chan bool //closing of this channel initiates the stop of the systemge component
	incomingConnectionsStopChannel       chan bool //closing of this channel indicates that the incoming connection handler has stopped
	allIncomingConnectionsStoppedChannel chan bool //closing of this channel indicates that all incoming connections have stopped
	messageHandlerChannel                chan func()

	incomingConnectionMutex  sync.RWMutex
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
		incomingConnections:  make(map[string]*incomingConnection),
		asyncMessageHandlers: asyncMessageHanlders,
		syncMessageHandlers:  syncMessageHandlers,
	}
}

func (systemgeServer *SystemgeServer) Start() error {
	if systemgeServer.config.TcpBufferBytes == 0 {
		systemgeServer.config.TcpBufferBytes = 1024 * 4
	}
	tcpServer, err := Tcp.NewServer(systemgeServer.config.ServerConfig)
	if err != nil {
		return Error.New("Failed to create tcp server", err)
	}
	if systemgeServer.config.IpRateLimiter != nil {
		systemgeServer.ipRateLimiter = Tools.NewIpRateLimiter(systemgeServer.config.IpRateLimiter)
	}

	systemgeServer.tcpServer = tcpServer
	systemgeServer.stopChannel = make(chan bool)
	systemgeServer.incomingConnectionsStopChannel = make(chan bool)
	systemgeServer.allIncomingConnectionsStoppedChannel = make(chan bool)

	if systemgeServer.config.ProcessAllMessagesSequentially {
		systemgeServer.messageHandlerChannel = make(chan func(), systemgeServer.config.ProcessAllMessagesSequentiallyChannelSize)
		if systemgeServer.config.ProcessAllMessagesSequentiallyChannelSize == 0 {
			go func() {
				for {
					select {
					case f := <-systemgeServer.messageHandlerChannel:
						f()
					case <-systemgeServer.stopChannel:
						return
					}
				}
			}()
		} else {
			go func() {
				for {
					select {
					case f := <-systemgeServer.messageHandlerChannel:
						if systemgeServer.errorLogger != nil && len(systemgeServer.messageHandlerChannel) >= systemgeServer.config.ProcessAllMessagesSequentiallyChannelSize-1 {
							systemgeServer.errorLogger.Log("ProcessAllMessagesSequentiallyChannelSize reached (increase ProcessAllMessagesSequentiallyChannelSize otherwise message order of arrival is not guaranteed)")
						}
						f()
					case <-systemgeServer.allIncomingConnectionsStoppedChannel:
						return
					}
				}
			}()
		}

	}
	go systemgeServer.handleIncomingConnections()
	return nil
}

// stopSystemgeServerComponent stops the systemge component.
// blocking until all goroutines associated with the systemge component have stopped.
func (systemge *SystemgeServer) Stop() error {
	close(systemge.stopChannel)

	if systemge.ipRateLimiter != nil {
		systemge.ipRateLimiter.Stop()
	}

	systemge.tcpServer.GetListener().Close()
	<-systemge.incomingConnectionsStopChannel

	systemge.incomingConnectionMutex.Lock()
	for _, incomingConnection := range systemge.incomingConnections {
		incomingConnection.netConn.Close()
		<-incomingConnection.stopChannel
	}
	systemge.incomingConnectionMutex.Unlock()
	close(systemge.allIncomingConnectionsStoppedChannel)
	return nil
}

func (systemge *SystemgeServer) GetName() string {
	return systemge.config.Name
}
