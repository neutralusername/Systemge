package Dashboard

import (
	"net/http"
	"runtime"
	"strings"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

type DashboardServer struct {
	name string

	statusMutex sync.Mutex
	status      int

	mutex sync.RWMutex

	clients map[string]*client

	frontendPath string

	config *Config.DashboardServer

	waitGroup sync.WaitGroup

	systemgeServer  *SystemgeServer.SystemgeServer
	httpServer      *HTTPServer.HTTPServer
	websocketServer *WebsocketServer.WebsocketServer

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer
}

func NewServer(name string, config *Config.DashboardServer) *DashboardServer {
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
	if config.SystemgeServerConfig.ListenerConfig == nil {
		panic("config.SystemgeServerConfig.ServerConfig is nil")
	}
	if config.SystemgeServerConfig.ListenerConfig.TcpServerConfig == nil {
		panic("config.SystemgeServerConfig.ServerConfig.ListenerConfig is nil")
	}
	if config.SystemgeServerConfig.ConnectionConfig == nil {
		panic("config.SystemgeServerConfig.ConnectionConfig is nil")
	}

	_, callerPath, _, _ := runtime.Caller(0)
	frontendPath := callerPath[:len(callerPath)-len("dashboardServer.go")] + "frontend/"
	Helpers.CreateFile(frontendPath+"configs.js",
		"export const WS_PORT = "+Helpers.Uint16ToString(config.WebsocketServerConfig.TcpServerConfig.Port)+";"+
			"export const WS_PATTERN = \""+config.WebsocketServerConfig.Pattern+"\";"+
			"export const MAX_CHART_ENTRIES = "+Helpers.Uint32ToString(config.MaxChartEntries)+";",
	)

	app := &DashboardServer{
		name:    name,
		mutex:   sync.RWMutex{},
		config:  config,
		clients: make(map[string]*client),

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

	app.systemgeServer = SystemgeServer.New(name+"_systemgeServer", app.config.SystemgeServerConfig, app.onSystemgeConnectHandler, app.onSystemgeDisconnectHandler)
	app.websocketServer = WebsocketServer.New(name+"_websocketServer", app.config.WebsocketServerConfig, map[string]WebsocketServer.MessageHandler{
		"start":   app.startHandler,
		"stop":    app.stopHandler,
		"command": app.commandHandler,
		"gc":      app.gcHandler,
	}, app.onWebsocketConnectHandler, nil)
	app.httpServer = HTTPServer.New(name+"_httpServer", app.config.HTTPServerConfig, nil)
	app.httpServer.AddRoute("/", HTTPServer.SendDirectory(app.frontendPath))

	return app
}

func (app *DashboardServer) Start() error {
	app.statusMutex.Lock()
	defer app.statusMutex.Unlock()
	if app.status == Status.STARTED {
		return Error.New("Already started", nil)
	}

	if err := app.systemgeServer.Start(); err != nil {
		return err
	}

	if err := app.websocketServer.Start(); err != nil {
		if err := app.systemgeServer.Stop(); err != nil {
			if app.errorLogger != nil {
				app.errorLogger.Log(Error.New("Failed to stop Systemge server after failed start", err).Error())
			}
		}
		return err
	}

	if err := app.httpServer.Start(); err != nil {
		if err := app.systemgeServer.Stop(); err != nil {
			if app.errorLogger != nil {
				app.errorLogger.Log(Error.New("Failed to stop Systemge server after failed start", err).Error())
			}
		}
		if err := app.websocketServer.Stop(); err != nil {
			if app.errorLogger != nil {
				app.errorLogger.Log(Error.New("Failed to stop Websocket server after failed start", err).Error())
			}
		}
		return err
	}

	app.status = Status.STARTED
	if app.config.StatusUpdateIntervalMs > 0 {
		app.waitGroup.Add(1)
		go app.statusUpdateRoutine()
	}
	if app.config.GoroutineUpdateIntervalMs > 0 {
		app.waitGroup.Add(1)
		go app.goroutineUpdateRoutine()
	}
	if app.config.HeapUpdateIntervalMs > 0 {
		app.waitGroup.Add(1)
		go app.heapUpdateRoutine()
	}
	if app.config.MetricsUpdateIntervalMs > 0 {
		app.waitGroup.Add(1)
		go app.metricsUpdateRoutine()
	}
	return nil
}

func (app *DashboardServer) Stop() error {
	app.statusMutex.Lock()
	defer app.statusMutex.Unlock()
	if app.status == Status.STOPPED {
		return Error.New("Already stopped", nil)
	}
	app.status = Status.PENDING
	app.waitGroup.Wait()

	if err := app.systemgeServer.Stop(); err != nil {
		if app.errorLogger != nil {
			app.errorLogger.Log(Error.New("Failed to stop Systemge server", err).Error())
		}
	}
	if err := app.websocketServer.Stop(); err != nil {
		if app.errorLogger != nil {
			app.errorLogger.Log(Error.New("Failed to stop Websocket server", err).Error())
		}
	}
	if err := app.httpServer.Stop(); err != nil {
		if app.errorLogger != nil {
			app.errorLogger.Log(Error.New("Failed to stop HTTP server", err).Error())
		}
	}

	app.status = Status.STOPPED
	return nil
}

func (app *DashboardServer) onSystemgeConnectHandler(connection SystemgeConnection.SystemgeConnection) error {
	response, err := connection.SyncRequestBlocking(Message.TOPIC_GET_INTRODUCTION, "")
	if err != nil {
		return err
	}
	client, err := unmarshalClient(response.GetPayload())
	if err != nil {
		return err
	}
	client.connection = connection

	app.mutex.Lock()
	app.registerModuleHttpHandlers(client)
	app.clients[client.Name] = client
	if client.Metrics == nil {
		client.Metrics = map[string]uint64{}
	}
	app.mutex.Unlock()
	app.websocketServer.Broadcast(Message.NewAsync("addModule", Helpers.JsonMarshal(client)))
	return nil
}

func (app *DashboardServer) onSystemgeDisconnectHandler(connection SystemgeConnection.SystemgeConnection) {
	app.mutex.Lock()
	if client, ok := app.clients[connection.GetName()]; ok {
		delete(app.clients, connection.GetName())
		app.unregisterModuleHttpHandlers(client.Name)
	}
	app.mutex.Unlock()
	app.websocketServer.Broadcast(Message.NewAsync("removeModule", connection.GetName()))
}

func (app *DashboardServer) registerModuleHttpHandlers(client *client) {
	app.httpServer.AddRoute("/"+client.Name, HTTPServer.SendDirectory(app.frontendPath))
	app.httpServer.AddRoute("/"+client.Name+"/command", func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		_, err := r.Body.Read(body)
		if err != nil {
			http.Error(w, Error.New("Failed to read body", err).Error(), http.StatusInternalServerError)
			return
		}
		args := strings.Split(string(body), " ")
		if len(args) == 0 {
			http.Error(w, "No command", http.StatusBadRequest)
			return
		}
		result, err := client.executeCommand(args[0], args[1:])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(result))
	})
	app.httpServer.AddRoute("/"+client.Name+"/command/", func(w http.ResponseWriter, r *http.Request) {
		args := r.URL.Path[len("/"+client.Name+"/command/"):]
		argsSplit := strings.Split(args, " ")
		if len(argsSplit) == 0 {
			http.Error(w, "No command", http.StatusBadRequest)
			return
		}
		result, err := client.executeCommand(argsSplit[0], argsSplit[1:])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(result))
	})
}

func (app *DashboardServer) unregisterModuleHttpHandlers(clientName string) {
	app.httpServer.RemoveRoute("/" + clientName)
	app.httpServer.RemoveRoute("/" + clientName + "/command/")
	app.httpServer.RemoveRoute("/" + clientName + "/command")
}
