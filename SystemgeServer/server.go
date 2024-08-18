package SystemgeServer

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeListener"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeServer struct {
	status      int
	statusMutex sync.Mutex

	config         *Config.SystemgeServer
	listener       *SystemgeListener.SystemgeListener
	messageHandler *SystemgeConnection.SystemgeMessageHandler

	clients            map[string]*SystemgeConnection.SystemgeConnection
	clientsMutex       sync.Mutex
	handlerStopChannel chan bool

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer
}

func New(config *Config.SystemgeServer, messageHandler *SystemgeConnection.SystemgeMessageHandler) *SystemgeServer {
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
	if messageHandler == nil {
		panic("messageHandler is nil")
	}
	server := &SystemgeServer{
		config:         config,
		messageHandler: messageHandler,
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
	handlerStopChannel := server.handlerStopChannel
	server.handlerStopChannel = nil
	<-handlerStopChannel
	server.listener.Close()
	server.listener = nil

	server.clientsMutex.Lock()
	for _, connection := range server.clients {
		connection.Close()
	}
	server.clientsMutex.Unlock()

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

func (server *SystemgeServer) handleConnections(stopChannel chan bool) {
	if server.infoLogger != nil {
		server.infoLogger.Log("connection handler started")
	}

	for server.handlerStopChannel == stopChannel {
		connection, err := server.listener.AcceptConnection(server.GetName(), server.config.ConnectionConfig, server.messageHandler)
		if err != nil {
			if server.warningLogger != nil {
				server.warningLogger.Log(Error.New("failed to accept connection", err).Error())
			}
			continue
		}
		if server.infoLogger != nil {
			server.infoLogger.Log("connection \"" + connection.GetName() + "\" accepted")
		}
		server.clientsMutex.Lock()
		server.clients[connection.GetName()] = connection
		server.clientsMutex.Unlock()
		go func() {
			<-connection.GetCloseChannel()
			server.clientsMutex.Lock()
			delete(server.clients, connection.GetName())
			server.clientsMutex.Unlock()
			if server.infoLogger != nil {
				server.infoLogger.Log("connection \"" + connection.GetName() + "\" closed")
			}
		}()
		if server.infoLogger != nil {
			server.infoLogger.Log("receiver for connection \"" + connection.GetName() + "\" started")
		}
	}
	close(stopChannel)

	if server.infoLogger != nil {
		server.infoLogger.Log("connection handler stopped")
	}
}
