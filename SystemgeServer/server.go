package SystemgeServer

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeListener"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
	"github.com/neutralusername/Systemge/SystemgeReceiver"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeServer struct {
	status      int
	statusMutex sync.RWMutex

	config   *Config.SystemgeServer
	listener *SystemgeListener.SystemgeListener

	messageHandler *SystemgeMessageHandler.SystemgeMessageHandler
	receiverConfig *Config.SystemgeReceiver

	clients            map[string]*SystemgeConnection.SystemgeConnection
	mutex              sync.Mutex
	handlerStopChannel chan bool

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer
}

func New(config *Config.SystemgeServer, receiverConfig *Config.SystemgeReceiver, messageHandler *SystemgeMessageHandler.SystemgeMessageHandler) *SystemgeServer {
	if config == nil {
		panic("config is nil")
	}
	if config.ConnectionConfig == nil {
		panic("config.ConnectionConfig is nil")
	}
	if config.ListenerConfig == nil {
		panic("listener is nil")
	}
	if config.ListenerConfig.TcpListenerConfig == nil {
		panic("listener.ListenerConfig is nil")
	}
	if messageHandler == nil && receiverConfig != nil {
		panic("receiverConfig is set but messageHandler is nil")
	}
	if messageHandler != nil && receiverConfig == nil {
		panic("messageHandler is set but receiverConfig is nil")
	}
	if config.ConnectionConfig.TcpBufferBytes == 0 {
		config.ConnectionConfig.TcpBufferBytes = 1024 * 4
	}
	server := &SystemgeServer{
		config:         config,
		messageHandler: messageHandler,
		receiverConfig: receiverConfig,
		clients:        make(map[string]*SystemgeConnection.SystemgeConnection),
	}
	if config.InfoLoggerPath != "" {
		server.infoLogger = Tools.NewLogger("[Info: \""+server.GetName()+"\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		server.warningLogger = Tools.NewLogger("[Warning: \""+server.GetName()+"\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		server.errorLogger = Tools.NewLogger("[Error: \""+server.GetName()+"\"] ", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		server.mailer = Tools.NewMailer(config.MailerConfig)
	}
	return server
}

func (server *SystemgeServer) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STOPPED {
		return Error.New("server is already started", nil)
	}
	server.status = Status.PENDING
	if server.infoLogger != nil {
		server.infoLogger.Log("starting server")
	}
	listener, err := SystemgeListener.New(server.config.ListenerConfig)
	if err != nil {
		server.status = Status.STOPPED
		return Error.New("failed to create listener", err)
	}
	server.listener = listener
	server.handlerStopChannel = make(chan bool)
	go server.handleConnections(server.handlerStopChannel)

	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log("server started")
	}
	server.status = Status.STARTED
	return nil
}

func (server *SystemgeServer) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STARTED {
		return Error.New("server is already stopped", nil)
	}
	server.status = Status.PENDING
	if server.infoLogger != nil {
		server.infoLogger.Log("stopping server")
	}

	server.listener.Close()
	handlerStopChannel := server.handlerStopChannel
	server.handlerStopChannel = nil
	<-handlerStopChannel
	server.listener = nil

	server.mutex.Lock()
	for _, connection := range server.clients {
		connection.Close()
	}
	server.mutex.Unlock()

	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log("server stopped")
	}
	server.status = Status.STOPPED
	return nil
}

func (server *SystemgeServer) GetName() string {
	return server.config.Name
}

func (server *SystemgeServer) GetStatus() int {
	return server.status
}

func (server *SystemgeServer) handleConnections(handlerStopChannel chan bool) {
	if server.infoLogger != nil {
		server.infoLogger.Log("connection handler started")
	}

	for server.handlerStopChannel == handlerStopChannel {
		connection, err := server.listener.AcceptConnection(server.GetName(), server.config.ConnectionConfig)
		if err != nil {
			if server.warningLogger != nil {
				server.warningLogger.Log(Error.New("failed to accept connection", err).Error())
			}
			continue
		}
		if server.infoLogger != nil {
			server.infoLogger.Log("connection \"" + connection.GetName() + "\" accepted")
		}
		if server.receiverConfig != nil && server.messageHandler != nil {
			// todo: add runtime funcs to do something with the receiver
			SystemgeReceiver.New(connection, server.receiverConfig, server.messageHandler)
		}

		server.mutex.Lock()
		if _, ok := server.clients[connection.GetName()]; ok {
			server.mutex.Unlock()
			if server.warningLogger != nil {
				server.warningLogger.Log("connection \"" + connection.GetName() + "\" already exists")
			}
			connection.Close()
			continue
		}
		server.clients[connection.GetName()] = connection
		server.mutex.Unlock()
		go func() {
			<-connection.GetCloseChannel()
			server.mutex.Lock()
			delete(server.clients, connection.GetName())
			server.mutex.Unlock()
			if server.infoLogger != nil {
				server.infoLogger.Log("connection \"" + connection.GetName() + "\" closed")
			}
		}()
		if server.infoLogger != nil {
			server.infoLogger.Log("receiver for connection \"" + connection.GetName() + "\" started")
		}
	}
	close(handlerStopChannel)

	if server.infoLogger != nil {
		server.infoLogger.Log("connection handler stopped")
	}
}
