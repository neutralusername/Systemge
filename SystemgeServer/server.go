package SystemgeServer

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeListener"
	"github.com/neutralusername/Systemge/TcpListener"
	"github.com/neutralusername/Systemge/Tools"
)

type OnConnectHandler func(SystemgeConnection.SystemgeConnection) error
type OnDisconnectHandler func(string, string)

type SystemgeServer struct {
	name string

	status      int
	statusMutex sync.RWMutex

	config   *Config.SystemgeServer
	listener SystemgeListener.SystemgeListener

	onConnectHandler    func(SystemgeConnection.SystemgeConnection) error
	onDisconnectHandler func(SystemgeConnection.SystemgeConnection)

	clients     map[string]SystemgeConnection.SystemgeConnection // name -> connection
	mutex       sync.Mutex
	stopChannel chan bool

	waitGroup sync.WaitGroup

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer
}

func New(name string, config *Config.SystemgeServer, onConnectHandler func(SystemgeConnection.SystemgeConnection) error, onDisconnectHandler func(SystemgeConnection.SystemgeConnection)) *SystemgeServer {
	if config == nil {
		panic("config is nil")
	}
	if config.ConnectionConfig == nil {
		panic("config.ConnectionConfig is nil")
	}
	if config.ListenerConfig == nil {
		panic("listener is nil")
	}
	if config.ListenerConfig.TcpServerConfig == nil {
		panic("listener.ListenerConfig is nil")
	}

	server := &SystemgeServer{
		name:                name,
		config:              config,
		onConnectHandler:    onConnectHandler,
		onDisconnectHandler: onDisconnectHandler,

		clients: make(map[string]SystemgeConnection.SystemgeConnection),
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
	listener, err := TcpListener.New(server.config.ListenerConfig)
	if err != nil {
		server.status = Status.STOPPED
		return Error.New("failed to create listener", err)
	}
	server.listener = listener
	server.stopChannel = make(chan bool)
	go server.handleConnections(server.stopChannel)

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

	close(server.stopChannel)
	server.listener.Close()

	server.mutex.Lock()
	for _, connection := range server.clients {
		connection.Close()
	}
	server.mutex.Unlock()

	server.waitGroup.Wait()
	server.stopChannel = nil
	server.listener = nil

	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log("server stopped")
	}
	server.status = Status.STOPPED
	return nil
}

func (server *SystemgeServer) GetName() string {
	return server.name
}

func (server *SystemgeServer) GetStatus() int {
	return server.status
}

func (server *SystemgeServer) GetBlacklist() *Tools.AccessControlList {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return nil
	}
	return server.listener.GetBlacklist()
}

func (server *SystemgeServer) GetWhitelist() *Tools.AccessControlList {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return nil
	}
	return server.listener.GetWhitelist()
}
