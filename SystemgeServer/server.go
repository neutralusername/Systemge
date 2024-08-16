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

type Client struct {
	connection *SystemgeConnection.SystemgeConnection
	receiver   *SystemgeReceiver.SystemgeReceiver
}

type SystemgeServer struct {
	status      int
	statusMutex sync.Mutex

	config         *Config.SystemgeServer
	listener       *SystemgeListener.SystemgeListener
	messageHandler *SystemgeMessageHandler.SystemgeMessageHandler

	clients            map[string]*Client
	clientsMutex       sync.Mutex
	handlerStopChannel chan bool

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer
}

func New(config *Config.SystemgeServer, messageHandler *SystemgeMessageHandler.SystemgeMessageHandler) *SystemgeServer {
	if config == nil {
		panic("config is nil")
	}
	if config.ConnectionConfig == nil {
		panic("config.ConnectionConfig is nil")
	}
	if config.ReceiverConfig == nil {
		panic("config.ReceiverConfig is nil")
	}
	if config.ListenerConfig == nil {
		panic("listener is nil")
	}
	if config.ListenerConfig.ListenerConfig == nil {
		panic("listener.ListenerConfig is nil")
	}
	if messageHandler == nil {
		panic("messageHandler is nil")
	}
	server := &SystemgeServer{
		config:         config,
		messageHandler: messageHandler,
		clients:        make(map[string]*Client),
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
	for _, client := range server.clients {
		client.receiver.Stop()
		client.connection.Close()
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
		connection, err := server.listener.AcceptConnection(server.GetName(), server.config.ConnectionConfig)
		if err != nil {
			if server.errorLogger != nil {
				server.errorLogger.Log("error accepting connection: " + err.Error())
			}
			continue
		}
		if server.infoLogger != nil {
			server.infoLogger.Log("connection accepted: " + connection.GetName())
		}
		receiver := SystemgeReceiver.New(server.config.ReceiverConfig, connection, server.messageHandler)
		if err := receiver.Start(); err != nil {
			connection.Close()
			if server.errorLogger != nil {
				server.errorLogger.Log("error starting receiver: " + err.Error())
			}
			continue
		}
		server.clientsMutex.Lock()
		server.clients[connection.GetName()] = &Client{
			connection: connection,
			receiver:   receiver,
		}
		server.clientsMutex.Unlock()
		if server.infoLogger != nil {
			server.infoLogger.Log("receiver started: " + connection.GetName())
		}
	}
	close(stopChannel)

	if server.infoLogger != nil {
		server.infoLogger.Log("connection handler stopped")
	}
}
