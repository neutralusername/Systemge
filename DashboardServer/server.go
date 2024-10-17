package DashboardServer

import (
	"runtime"
	"sync"

	"github.com/neutralusername/systemge/Commands"
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/DashboardHelpers"
	"github.com/neutralusername/systemge/Event"
	Server1 "github.com/neutralusername/systemge/Server"
	"github.com/neutralusername/systemge/WebsocketServer"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/httpServer"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/tools"
)

type Server struct {
	config *configs.DashboardServer

	name string

	statusMutex sync.Mutex
	status      int

	frontendPath              string
	dashboardCommandHandlers  Commands.Handlers
	dashboardWebsocketClients map[*WebsocketServer.WebsocketConnection]bool // websocketClient -> true (websocketClients that are currently on the dashboard page)

	waitGroup sync.WaitGroup
	mutex     sync.RWMutex

	dashboardClient          *DashboardHelpers.DashboardClient
	connectedClients         map[string]*connectedClient                     // name/location -> connectedClient
	websocketClientLocations map[*WebsocketServer.WebsocketConnection]string // websocketId -> location ("/" == dashboard/landing page) ("" == no location)

	systemgeServer  *Server1.Server
	httpServer      *httpServer.HTTPServer
	websocketServer *WebsocketServer.WebsocketServer

	responseMessageCache      map[string]*DashboardHelpers.ResponseMessage
	responseMessageCacheOrder []*DashboardHelpers.ResponseMessage // has the purpose, to easily remove the oldest response message without iterating over the whole map

	infoLogger    *tools.Logger
	warningLogger *tools.Logger
	errorLogger   *tools.Logger
	mailer        *tools.Mailer
}

func New(name string, config *configs.DashboardServer, whitelist *tools.AccessControlList, blacklist *tools.AccessControlList) *Server {
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
	if config.MaxEntriesPerMetrics <= 0 {
		config.MaxEntriesPerMetrics = 100
	}
	if config.ResponseMessageCacheSize <= 0 {
		config.ResponseMessageCacheSize = 100
	}
	if config.FrontendHeartbeatIntervalMs == 0 {
		config.FrontendHeartbeatIntervalMs = 1000 * 60
	}

	_, callerPath, _, _ := runtime.Caller(0)
	frontendPath := callerPath[:len(callerPath)-len("server.go")] + "frontend/"
	helpers.CreateFile(frontendPath+"configs.js", "export const configs = "+helpers.JsonMarshal(map[string]interface{}{
		"WS_PORT":                     config.WebsocketServerConfig.TcpServerConfig.Port,
		"WS_PATTERN":                  config.WebsocketServerConfig.Pattern,
		"FRONTEND_HEARTBEAT_INTERVAL": config.FrontendHeartbeatIntervalMs,
		"RESPONSE_MESSAGE_CACHE_SIZE": config.ResponseMessageCacheSize,
		"MAX_ENTRIES_PER_METRICS":     config.MaxEntriesPerMetrics,
	}))

	server := &Server{
		name:                      name,
		mutex:                     sync.RWMutex{},
		config:                    config,
		connectedClients:          map[string]*connectedClient{},
		websocketClientLocations:  map[*WebsocketServer.WebsocketConnection]string{},
		dashboardWebsocketClients: map[*WebsocketServer.WebsocketConnection]bool{},
		responseMessageCache:      map[string]*DashboardHelpers.ResponseMessage{},
		responseMessageCacheOrder: []*DashboardHelpers.ResponseMessage{},
		frontendPath:              frontendPath,
	}

	if config.InfoLoggerPath != "" {
		server.infoLogger = tools.NewLogger("[Info: \""+name+"\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		server.warningLogger = tools.NewLogger("[Warning: \""+name+"\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		server.errorLogger = tools.NewLogger("[Error: \""+name+"\"] ", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		server.mailer = tools.NewMailer(config.MailerConfig)
	}

	server.systemgeServer = Server1.New(name+"_systemgeServer",
		server.config.SystemgeServerConfig,
		whitelist, blacklist,
		server.onSystemgeConnectHandler, server.onSystemgeDisconnectHandler,
	)
	server.websocketServer = WebsocketServer.New(name+"_websocketServer",
		server.config.WebsocketServerConfig,
		whitelist, blacklist,
		map[string]WebsocketServer.MessageHandler{
			DashboardHelpers.TOPIC_PAGE_REQUEST: server.pageRequestHandler,
			DashboardHelpers.TOPIC_CHANGE_PAGE:  server.changePageHandler,
		},
		server.onWebsocketConnectHandler, server.onWebsocketDisconnectHandler,
	)
	server.httpServer = httpServer.New(name+"_httpServer",
		server.config.HTTPServerConfig,
		whitelist, blacklist,
		nil,
	)
	server.httpServer.AddRoute("/", httpServer.SendDirectory(server.frontendPath))

	server.dashboardCommandHandlers = server.GetDefaultCommands()
	return server
}

func (server *Server) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status == status.Started {
		return Event.New("Already started", nil)
	}
	if err := server.systemgeServer.Start(); err != nil {
		return err
	}
	if err := server.websocketServer.Start(); err != nil {
		if err := server.systemgeServer.Stop(); err != nil {
			if server.errorLogger != nil {
				server.errorLogger.Log(Event.New("Failed to stop Systemge server after failed start", err).Error())
			}
		}
		return err
	}
	if err := server.httpServer.Start(); err != nil {
		if err := server.systemgeServer.Stop(); err != nil {
			if server.errorLogger != nil {
				server.errorLogger.Log(Event.New("Failed to stop Systemge server after failed start", err).Error())
			}
		}
		if err := server.websocketServer.Stop(); err != nil {
			if server.errorLogger != nil {
				server.errorLogger.Log(Event.New("Failed to stop Websocket server after failed start", err).Error())
			}
		}
		return err
	}
	server.status = status.Started

	if server.config.UpdateIntervalMs > 0 {
		server.waitGroup.Add(1)
		go server.updateRoutine()
	}

	return nil
}

func (server *Server) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status == status.Stopped {
		return Event.New("Already stopped", nil)
	}
	server.status = status.Pending
	server.waitGroup.Wait()
	if err := server.systemgeServer.Stop(); err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log(Event.New("Failed to stop Systemge server", err).Error())
		}
	}
	if err := server.websocketServer.Stop(); err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log(Event.New("Failed to stop Websocket server", err).Error())
		}
	}
	if err := server.httpServer.Stop(); err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log(Event.New("Failed to stop HTTP server", err).Error())
		}
	}
	server.status = status.Stopped
	return nil
}

func (server *Server) GetWebsocketClientIdsOnPage(page string) []string {
	server.mutex.RLock()
	defer server.mutex.RUnlock()
	clients := make([]string, 0)
	switch page {
	case "":
	case DashboardHelpers.DASHBOARD_CLIENT_NAME:
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
