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
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

type DashboardServer struct {
	closed bool

	clients map[string]*client

	mutex  sync.RWMutex
	config *Config.DashboardServer

	systemgeServer  *SystemgeServer.SystemgeServer
	httpServer      *HTTPServer.HTTPServer
	websocketServer *WebsocketServer.WebsocketServer

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer
}

func NewServer(config *Config.DashboardServer) *DashboardServer {
	if config == nil {
		panic("config is nil")
	}
	if config.HTTPServerConfig == nil {
		panic("config.HTTPServerConfig is nil")
	}
	if config.HTTPServerConfig.TcpListenerConfig == nil {
		panic("config.HTTPServerConfig.ServerConfig is nil")
	}
	if config.WebsocketServerConfig == nil {
		panic("config.WebsocketServerConfig is nil")
	}
	if config.WebsocketServerConfig.Pattern == "" {
		panic("config.WebsocketServerConfig.Pattern is empty")
	}
	if config.WebsocketServerConfig.TcpListenerConfig == nil {
		panic("config.WebsocketServerConfig.ServerConfig is nil")
	}
	if config.SystemgeServerConfig == nil {
		panic("config.SystemgeServerConfig is nil")
	}
	if config.SystemgeServerConfig.ListenerConfig == nil {
		panic("config.SystemgeServerConfig.ServerConfig is nil")
	}
	if config.SystemgeServerConfig.ListenerConfig.TcpListenerConfig == nil {
		panic("config.SystemgeServerConfig.ServerConfig.ListenerConfig is nil")
	}
	if config.SystemgeServerConfig.ConnectionConfig == nil {
		panic("config.SystemgeServerConfig.ConnectionConfig is nil")
	}
	if config.SystemgeServerConfig.ConnectionConfig.TcpBufferBytes == 0 {
		config.SystemgeServerConfig.ConnectionConfig.TcpBufferBytes = 1024 * 4
	}
	app := &DashboardServer{
		mutex:   sync.RWMutex{},
		config:  config,
		clients: make(map[string]*client),

		infoLogger:    Tools.NewLogger("[Info: \"Dashboard\"]", config.InfoLoggerPath),
		warningLogger: Tools.NewLogger("[Warning: \"Dashboard\"]", config.WarningLoggerPath),
		errorLogger:   Tools.NewLogger("[Error: \"Dashboard\"]", config.ErrorLoggerPath),
		mailer:        Tools.NewMailer(config.MailerConfig),
	}
	app.httpServer = HTTPServer.New(config.HTTPServerConfig, nil)
	_, callerPath, _, _ := runtime.Caller(0)
	frontendPath := callerPath[:len(callerPath)-len("dashboardServer.go")] + "frontend/"
	Helpers.CreateFile(frontendPath+"configs.js", "export const WS_PORT = "+Helpers.Uint16ToString(config.WebsocketServerConfig.TcpListenerConfig.Port)+";export const WS_PATTERN = \""+config.WebsocketServerConfig.Pattern+"\";")
	app.httpServer.AddRoute("/", HTTPServer.SendDirectory(frontendPath))

	app.websocketServer = WebsocketServer.New(config.WebsocketServerConfig, map[string]WebsocketServer.MessageHandler{
		"start":   app.startHandler,
		"stop":    app.stopHandler,
		"command": app.commandHandler,
		"gc":      app.gcHandler,
	}, app.onWebsocketConnectHandler, nil)
	app.systemgeServer = SystemgeServer.New(config.SystemgeServerConfig, app.onSystemgeConnectHandler, app.onSystemgeDisconnectHandler)

	err := app.httpServer.Start()
	if err != nil {
		panic(err)
	}
	err = app.websocketServer.Start()
	if err != nil {
		panic(err)
	}
	err = app.systemgeServer.Start()
	if err != nil {
		panic(err)
	}

	if app.config.GoroutineUpdateIntervalMs > 0 {
		go app.goroutineUpdateRoutine()
	}
	if app.config.StatusUpdateIntervalMs > 0 {
		go app.statusUpdateRoutine()
	}
	if app.config.HeapUpdateIntervalMs > 0 {
		go app.heapUpdateRoutine()
	}
	if app.config.MetricsUpdateIntervalMs > 0 {
		go app.metricsUpdateRoutine()
	}

	return app
}

func (app *DashboardServer) Close() {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.closed {
		return
	}
	app.closed = true
	app.httpServer.Stop()
	app.websocketServer.Stop()
	app.systemgeServer.Stop()
}

func (app *DashboardServer) onSystemgeConnectHandler(connection *SystemgeConnection.SystemgeConnection) error {
	response, err := connection.SyncRequest(Message.TOPIC_GET_INTRODUCTION, "")
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
	app.mutex.Unlock()
	app.websocketServer.Broadcast(Message.NewAsync("addModule", Helpers.JsonMarshal(client)))
	return nil
}

func (app *DashboardServer) onSystemgeDisconnectHandler(connection *SystemgeConnection.SystemgeConnection) {
	app.mutex.Lock()
	if client, ok := app.clients[connection.GetName()]; ok {
		delete(app.clients, connection.GetName())
		app.unregisterModuleHttpHandlers(client.Name)
	}
	app.mutex.Unlock()
	app.websocketServer.Broadcast(Message.NewAsync("removeModule", connection.GetName()))
}

func (app *DashboardServer) registerModuleHttpHandlers(client *client) {
	_, filePath, _, _ := runtime.Caller(0)

	app.httpServer.AddRoute("/"+client.Name, func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/"+client.Name, http.FileServer(http.Dir(filePath[:len(filePath)-len("dasbboardServer.go")]+"frontend"))).ServeHTTP(w, r)
	})
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
}
