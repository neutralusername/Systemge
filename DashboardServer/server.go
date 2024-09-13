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

	connectedClients         map[string]*connectedClient
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
		"MAX_CHART_ENTRIES":           config.MaxChartEntries,
		"FRONTEND_HEARTBEAT_INTERVAL": config.FrontendHeartbeatIntervalMs,
	}))

	app := &Server{
		name:                     name,
		mutex:                    sync.RWMutex{},
		config:                   config,
		connectedClients:         make(map[string]*connectedClient),
		websocketClientLocations: make(map[*WebsocketServer.WebsocketClient]string),

		frontendPath: frontendPath,
	}

	if config.InfoLoggerPath != "" {
		app.infoLogger = Tools.NewLogger("[Info: \"DashboardServer\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		app.warningLogger = Tools.NewLogger("[Warning: \"DashboardServer\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		app.errorLogger = Tools.NewLogger("[Error: \"DashboardServer\"] ", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		app.mailer = Tools.NewMailer(config.MailerConfig)
	}

	app.systemgeServer = SystemgeServer.New(name+"_systemgeServer",
		app.config.SystemgeServerConfig,
		whitelist, blacklist,
		app.onSystemgeConnectHandler, app.onSystemgeDisconnectHandler,
	)
	app.websocketServer = WebsocketServer.New(name+"_websocketServer",
		app.config.WebsocketServerConfig,
		whitelist, blacklist,
		map[string]WebsocketServer.MessageHandler{
			DashboardHelpers.TOPIC_PAGE_REQUEST: app.pageRequestHandler,
			DashboardHelpers.TOPIC_CHANGE_PAGE:  app.handleChangePage,
		},
		app.onWebsocketConnectHandler, app.onWebsocketDisconnectHandler,
	)
	app.httpServer = HTTPServer.New(name+"_httpServer",
		app.config.HTTPServerConfig,
		whitelist, blacklist,
		nil,
	)
	app.httpServer.AddRoute("/", HTTPServer.SendDirectory(app.frontendPath))

	app.dashboardCommandHandlers = Commands.Handlers{
		"disconnectClient": func(args []string) (string, error) {
			if len(args) == 0 {
				return "", Error.New("No client name", nil)
			}
			if err := app.DisconnectClient(args[0]); err != nil {
				return "", err
			}
			return "success", nil
		},
	}
	if app.config.DashboardSystemgeCommands {
		systemgeDefaultCommands := app.systemgeServer.GetDefaultCommands()
		for command, handler := range systemgeDefaultCommands {
			app.dashboardCommandHandlers["systemgeServer_"+command] = handler
		}
	}
	if app.config.DashboardWebsocketCommands {
		httpDefaultCommands := app.httpServer.GetDefaultCommands()
		for command, handler := range httpDefaultCommands {
			app.dashboardCommandHandlers["httpServer_"+command] = handler
		}
	}
	if app.config.DashboardHttpCommands {
		webSocketDefaultCommands := app.websocketServer.GetDefaultCommands()
		for command, handler := range webSocketDefaultCommands {
			app.dashboardCommandHandlers["websocketServer_"+command] = handler
		}
	}

	return app
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
