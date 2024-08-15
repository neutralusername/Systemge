package Dashboard

import (
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Module"
	"github.com/neutralusername/Systemge/SystemgeServer"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

type App struct {
	started bool

	serviceModules map[string]Module.ServiceModule
	modules        map[string]Module.Module
	mutex          sync.RWMutex
	config         *Config.Dashboard

	systemgeServer  *SystemgeServer.SystemgeServer
	httpServer      *HTTPServer.HTTPServer
	websocketServer *WebsocketServer.WebsocketServer

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer
}

func New(config *Config.Dashboard) *App {
	if config == nil {
		panic("config is nil")
	}
	if config.HTTPServerConfig == nil {
		panic("config.HTTPServerConfig is nil")
	}
	if config.HTTPServerConfig.ServerConfig == nil {
		panic("config.HTTPServerConfig.ServerConfig is nil")
	}
	if config.WebsocketServerConfig == nil {
		panic("config.WebsocketServerConfig is nil")
	}
	if config.WebsocketServerConfig.ServerConfig == nil {
		panic("config.WebsocketServerConfig.ServerConfig is nil")
	}
	if config.SystemgeServerConfig == nil {
		panic("config.SystemgeServerConfig is nil")
	}
	if config.SystemgeServerConfig.ServerConfig == nil {
		panic("config.SystemgeServerConfig.ServerConfig is nil")
	}
	app := &App{
		serviceModules: make(map[string]Module.ServiceModule),
		modules:        make(map[string]Module.Module),
		mutex:          sync.RWMutex{},
		config:         config,

		infoLogger:    Tools.NewLogger("[Info: \"Dashboard\"]", config.InfoLoggerPath),
		warningLogger: Tools.NewLogger("[Warning: \"Dashboard\"]", config.WarningLoggerPath),
		errorLogger:   Tools.NewLogger("[Error: \"Dashboard\"]", config.ErrorLoggerPath),
		mailer:        Tools.NewMailer(config.MailerConfig),
	}
	app.httpServer = HTTPServer.New(config.HTTPServerConfig, nil)
	_, callerPath, _, _ := runtime.Caller(0)
	frontendPath := callerPath[:len(callerPath)-len("app.go")] + "frontend/"
	Helpers.CreateFile(frontendPath+"configs.js", "export const WS_PORT = "+Helpers.Uint16ToString(config.WebsocketServerConfig.ServerConfig.Port)+";export const WS_PATTERN = \""+config.WebsocketServerConfig.Pattern+"\";")
	app.httpServer.AddRoute("/", HTTPServer.SendDirectory(frontendPath))

	app.websocketServer = WebsocketServer.New(config.WebsocketServerConfig, app.GetWebsocketMessageHandlers(), app.OnConnectHandler, app.OnDisconnectHandler)
	app.systemgeServer = SystemgeServer.New(config.SystemgeServerConfig, map[string]SystemgeServer.AsyncMessageHandler{}, map[string]SystemgeServer.SyncMessageHandler{})
	return app
}

func (app *App) Start() error {
	err := app.httpServer.Start()
	if err != nil {
		return err
	}
	err = app.websocketServer.Start()
	if err != nil {
		app.httpServer.Stop()
		return err
	}
	err = app.systemgeServer.Start()
	if err != nil {
		app.httpServer.Stop()
		app.websocketServer.Stop()
		return err
	}

	app.started = true

	if app.config.GoroutineUpdateIntervalMs > 0 {
		go app.goroutineUpdateRoutine()
	}
	if app.config.ServiceStatusUpdateIntervalMs > 0 {
		go app.serviceStatusUpdateRoutine()
	}
	if app.config.HeapUpdateIntervalMs > 0 {
		go app.heapUpdateRoutine()
	}

	return nil
}

func (app *App) Stop() error {
	app.started = false
	app.httpServer.Stop()
	app.websocketServer.Stop()
	app.systemgeServer.Stop()
	return nil
}

func (app *App) registerModuleHttpHandlers(module Module.Module) {
	_, filePath, _, _ := runtime.Caller(0)

	app.httpServer.AddRoute("/"+module.GetName(), func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/"+module.GetName(), http.FileServer(http.Dir(filePath[:len(filePath)-len("lifecycle.go")]+"frontend"))).ServeHTTP(w, r)
	})
	app.httpServer.AddRoute("/"+module.GetName()+"/command/", func(w http.ResponseWriter, r *http.Request) {
		args := r.URL.Path[len("/"+module.GetName()+"/command/"):]
		argsSplit := strings.Split(args, " ")
		if len(argsSplit) == 0 {
			http.Error(w, "No command", http.StatusBadRequest)
			return
		}
		result, err := app.nodeCommand(&Command{Name: module.GetName(), Command: argsSplit[0], Args: argsSplit[1:]})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(result))
	})
}

func (app *App) unregisterNodeHttpHandlers(module Module.Module) {
	app.httpServer.RemoveRoute("/" + module.GetName())
	app.httpServer.RemoveRoute("/" + module.GetName() + "/command/")
}

func (app *App) serviceStatusUpdateRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.serviceModules {
			statusUpdateJson := Helpers.JsonMarshal(newServiceStatus(node))
			go app.websocketServer.Broadcast(Message.NewAsync("nodeStatus", statusUpdateJson))
			if infoLogger := app.infoLogger; infoLogger != nil {
				infoLogger.Log("status update routine: \"" + statusUpdateJson + "\"")
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.ServiceStatusUpdateIntervalMs) * time.Millisecond)
	}
}

func (app *App) goroutineUpdateRoutine() {
	for app.started {
		goroutineCount := runtime.NumGoroutine()
		go app.websocketServer.Broadcast(Message.NewAsync("goroutineCount", strconv.Itoa(goroutineCount)))
		if infoLogger := app.infoLogger; infoLogger != nil {
			infoLogger.Log("goroutine update routine: \"" + strconv.Itoa(goroutineCount) + "\"")
		}
		time.Sleep(time.Duration(app.config.GoroutineUpdateIntervalMs) * time.Millisecond)
	}
}

func (app *App) heapUpdateRoutine() {
	for app.started {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		heapSize := strconv.FormatUint(memStats.HeapSys, 10)
		go app.websocketServer.Broadcast(Message.NewAsync("heapStatus", heapSize))
		if infoLogger := app.infoLogger; infoLogger != nil {
			infoLogger.Log("heap update routine: \"" + heapSize + "\"")
		}
		time.Sleep(time.Duration(app.config.HeapUpdateIntervalMs) * time.Millisecond)
	}
}
