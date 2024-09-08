package SystemgeServer

import (
	"encoding/json"
	"sync"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeListener"
	"github.com/neutralusername/Systemge/TcpSystemgeListener"
	"github.com/neutralusername/Systemge/Tools"
)

type OnConnectHandler func(SystemgeConnection.SystemgeConnection) error
type OnDisconnectHandler func(string, string)

type SystemgeServer struct {
	name string

	status      int
	statusMutex sync.RWMutex

	whitelist *Tools.AccessControlList
	blacklist *Tools.AccessControlList

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

func New(name string, config *Config.SystemgeServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, onConnectHandler func(SystemgeConnection.SystemgeConnection) error, onDisconnectHandler func(SystemgeConnection.SystemgeConnection)) *SystemgeServer {
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

		whitelist: whitelist,
		blacklist: blacklist,

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
	listener, err := TcpSystemgeListener.New(server.config.ListenerConfig, server.whitelist, server.blacklist)
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

	// rare race condition that causes panic code line below - panic: sync: WaitGroup is reused before previous Wait has returned
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
	return server.blacklist
}

func (server *SystemgeServer) GetWhitelist() *Tools.AccessControlList {
	return server.whitelist
}

func (server *SystemgeServer) GetDefaultCommands() Commands.Handlers {
	serverCommands := Commands.Handlers{}
	blacklistCommands := server.GetBlacklist().GetDefaultCommands()
	whitelistCommands := server.GetWhitelist().GetDefaultCommands()
	commands := Commands.Handlers{}
	for key, value := range blacklistCommands {
		commands["blacklist_"+key] = value
	}
	for key, value := range whitelistCommands {
		commands["whitelist_"+key] = value
	}
	serverCommands["start"] = func(args []string) (string, error) {
		err := server.Start()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	serverCommands["stop"] = func(args []string) (string, error) {
		err := server.Stop()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	serverCommands["removeConnection"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("expected 1 argument", nil)
		}
		err := server.RemoveConnection(args[0])
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	serverCommands["getConnectionNamesAndAddresses"] = func(args []string) (string, error) {
		names := server.GetConnectionNamesAndAddresses()
		json, err := json.Marshal(names)
		if err != nil {
			return "", Error.New("failed to marshal map to json", err)
		}
		return string(json), nil
	}
	serverCommands["getConnectionCount"] = func(args []string) (string, error) {
		return Helpers.IntToString(server.GetConnectionCount()), nil
	}
	serverCommands["asyncMessage"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", Error.New("expected at least 2 arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		clientNames := args[2:]
		err := server.AsyncMessage(topic, payload, clientNames...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	serverCommands["syncRequest"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", Error.New("expected at least 2 arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		clientNames := args[2:]
		messages, err := server.SyncRequestBlocking(topic, payload, clientNames...)
		if err != nil {
			return "", err
		}
		json, err := json.Marshal(messages)
		if err != nil {
			return "", Error.New("failed to marshal messages to json", err)
		}
		return string(json), nil
	}
	serverCommands["getMetrics"] = func(args []string) (string, error) {
		metrics := server.GetMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Error.New("failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	serverCommands["retrieveMetrics"] = func(args []string) (string, error) {
		metrics := server.RetrieveMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Error.New("failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	return serverCommands
}
