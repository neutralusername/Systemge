package DashboardServer

import (
	"runtime"
	"sync"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeServer"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

type Server struct {
	config *Config.DashboardServer

	name string

	statusMutex sync.Mutex
	status      int

	frontendPath              string
	dashboardCommandHandlers  Commands.Handlers
	dashboardWebsocketClients map[*WebsocketServer.WebsocketClient]bool // websocketClient -> true (websocketClients that are currently on the dashboard page)

	waitGroup sync.WaitGroup
	mutex     sync.RWMutex

	dashboardClient          *dashboardClient
	connectedClients         map[string]*connectedClient                 // name/location -> connectedClient
	websocketClientLocations map[*WebsocketServer.WebsocketClient]string // websocketId -> location ("/" == dashboard/landing page) ("" == no location)

	systemgeServer  *SystemgeServer.SystemgeServer
	httpServer      *HTTPServer.HTTPServer
	websocketServer *WebsocketServer.WebsocketServer

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer
}

func New(name string, config *Config.DashboardServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList) *Server {
	if config == nil {
		panic("config is nil")
	}
	if config.MaxMetricsCacheValues < 0 {
		panic("config.MaxMetricsCacheValues must be >= 0")
	}
	if config.HTTPServerConfig == nil {
		panic("config.HTTPServerConfig is nil")
	}
	if config.HTTPServerConfig.TcpServerConfig == nil {
		panic("config.HTTPServerConfig.ServerConfig is nil")
	}
	if config.WebsocketServerConfig == nil {
		panic("config.WebsocketServerConfig is nil")
	}
	if config.WebsocketServerConfig.Pattern == "" {
		panic("config.WebsocketServerConfig.Pattern is empty")
	}
	if config.WebsocketServerConfig.TcpServerConfig == nil {
		panic("config.WebsocketServerConfig.ServerConfig is nil")
	}
	if config.SystemgeServerConfig == nil {
		panic("config.SystemgeServerConfig is nil")
	}
	if config.SystemgeServerConfig.TcpSystemgeListenerConfig == nil {
		panic("config.SystemgeServerConfig.ServerConfig is nil")
	}
	if config.SystemgeServerConfig.TcpSystemgeListenerConfig.TcpServerConfig == nil {
		panic("config.SystemgeServerConfig.ServerConfig.ListenerConfig is nil")
	}
	if config.SystemgeServerConfig.TcpSystemgeConnectionConfig == nil {
		panic("config.SystemgeServerConfig.ConnectionConfig is nil")
	}

	_, callerPath, _, _ := runtime.Caller(0)
	frontendPath := callerPath[:len(callerPath)-len("server.go")] + "frontend/"
	Helpers.CreateFile(frontendPath+"configs.js", "export const configs = "+Helpers.JsonMarshal(map[string]interface{}{
		"WS_PORT":                     config.WebsocketServerConfig.TcpServerConfig.Port,
		"WS_PATTERN":                  config.WebsocketServerConfig.Pattern,
		"FRONTEND_HEARTBEAT_INTERVAL": config.FrontendHeartbeatIntervalMs,
	}))

	server := &Server{
		name:                      name,
		mutex:                     sync.RWMutex{},
		config:                    config,
		connectedClients:          map[string]*connectedClient{},
		websocketClientLocations:  map[*WebsocketServer.WebsocketClient]string{},
		dashboardWebsocketClients: map[*WebsocketServer.WebsocketClient]bool{},
		frontendPath:              frontendPath,
	}

	if config.InfoLoggerPath != "" {
		server.infoLogger = Tools.NewLogger("[Info: \""+name+"\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		server.warningLogger = Tools.NewLogger("[Warning: \""+name+"\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		server.errorLogger = Tools.NewLogger("[Error: \""+name+"\"] ", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		server.mailer = Tools.NewMailer(config.MailerConfig)
	}

	server.systemgeServer = SystemgeServer.New(name+"_systemgeServer",
		server.config.SystemgeServerConfig,
		whitelist, blacklist,
		server.onSystemgeConnectHandler, server.onSystemgeDisconnectHandler,
	)
	server.websocketServer = WebsocketServer.New(name+"_websocketServer",
		server.config.WebsocketServerConfig,
		whitelist, blacklist,
		map[string]WebsocketServer.MessageHandler{
			DashboardHelpers.TOPIC_PAGE_REQUEST: server.pageRequestHandler,
			DashboardHelpers.TOPIC_CHANGE_PAGE:  server.handleChangePage,
		},
		server.onWebsocketConnectHandler, server.onWebsocketDisconnectHandler,
	)
	server.httpServer = HTTPServer.New(name+"_httpServer",
		server.config.HTTPServerConfig,
		whitelist, blacklist,
		nil,
	)
	server.httpServer.AddRoute("/", HTTPServer.SendDirectory(server.frontendPath))

	server.dashboardCommandHandlers = Commands.Handlers{
		"disconnectClient": func(args []string) (string, error) {
			if len(args) == 0 {
				return "", Error.New("No client name", nil)
			}
			if err := server.DisconnectClient(args[0]); err != nil {
				return "", err
			}
			return "success", nil
		},
	}
	if server.config.DashboardSystemgeCommands {
		systemgeDefaultCommands := server.systemgeServer.GetDefaultCommands()
		for command, handler := range systemgeDefaultCommands {
			server.dashboardCommandHandlers["systemgeServer_"+command] = handler
		}
	}
	if server.config.DashboardWebsocketCommands {
		httpDefaultCommands := server.httpServer.GetDefaultCommands()
		for command, handler := range httpDefaultCommands {
			server.dashboardCommandHandlers["httpServer_"+command] = handler
		}
	}
	if server.config.DashboardHttpCommands {
		webSocketDefaultCommands := server.websocketServer.GetDefaultCommands()
		for command, handler := range webSocketDefaultCommands {
			server.dashboardCommandHandlers["websocketServer_"+command] = handler
		}
	}
	server.dashboardClient = &dashboardClient{
		Name:           DASHBOARD_CLIENT_NAME,
		ClientStatuses: map[string]int{},
		Commands:       server.dashboardCommandHandlers.GetKeyBoolMap(),
		Metrics:        map[string]map[string][]*DashboardHelpers.MetricsEntry{},
	}
	return server
}

func (server *Server) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status == Status.STARTED {
		return Error.New("Already started", nil)
	}
	if err := server.systemgeServer.Start(); err != nil {
		return err
	}
	if err := server.websocketServer.Start(); err != nil {
		if err := server.systemgeServer.Stop(); err != nil {
			if server.errorLogger != nil {
				server.errorLogger.Log(Error.New("Failed to stop Systemge server after failed start", err).Error())
			}
		}
		return err
	}
	if err := server.httpServer.Start(); err != nil {
		if err := server.systemgeServer.Stop(); err != nil {
			if server.errorLogger != nil {
				server.errorLogger.Log(Error.New("Failed to stop Systemge server after failed start", err).Error())
			}
		}
		if err := server.websocketServer.Stop(); err != nil {
			if server.errorLogger != nil {
				server.errorLogger.Log(Error.New("Failed to stop Websocket server after failed start", err).Error())
			}
		}
		return err
	}
	server.status = Status.STARTED

	if server.config.UpdateIntervalMs > 0 {
		server.waitGroup.Add(1)
		go server.updateRoutine()
	}

	return nil
}

func (server *Server) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status == Status.STOPPED {
		return Error.New("Already stopped", nil)
	}
	server.status = Status.PENDING
	server.waitGroup.Wait()
	if err := server.systemgeServer.Stop(); err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log(Error.New("Failed to stop Systemge server", err).Error())
		}
	}
	if err := server.websocketServer.Stop(); err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log(Error.New("Failed to stop Websocket server", err).Error())
		}
	}
	if err := server.httpServer.Stop(); err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log(Error.New("Failed to stop HTTP server", err).Error())
		}
	}
	server.status = Status.STOPPED
	return nil
}

func (server *Server) GetWebsocketClientIdsOnPage(page string) []string {
	server.mutex.RLock()
	defer server.mutex.RUnlock()
	clients := make([]string, 0)
	switch page {
	case "":
	case DASHBOARD_CLIENT_NAME:
		for client := range server.dashboardWebsocketClients {
			clients = append(clients, client.GetId())
		}
	default:
		connectedClient := server.connectedClients[page]
		if connectedClient != nil {
			for client := range connectedClient.websocketClients {
				clients = append(clients, client.GetId())
			}
		}
	}
	return clients
}
