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

func New(dashboardConfig *Config.Dashboard) *App {
	app := &App{
		serviceModules: make(map[string]Module.ServiceModule),
		modules:        make(map[string]Module.Module),
		mutex:          sync.RWMutex{},
		config:         dashboardConfig,

		httpServer:      nil,
		websocketServer: nil,

		infoLogger:    Tools.NewLogger("[Info: \"Dashboard\"]", dashboardConfig.InfoLoggerPath),
		warningLogger: Tools.NewLogger("[Warning: \"Dashboard\"]", dashboardConfig.WarningLoggerPath),
		errorLogger:   Tools.NewLogger("[Error: \"Dashboard\"]", dashboardConfig.ErrorLoggerPath),
		mailer:        Tools.NewMailer(dashboardConfig.Mailer),
	}
	app.httpServer = HTTPServer.New(&Config.HTTPServer{
		ServerConfig: dashboardConfig.ServerConfig,
	}, nil)

	app.websocketServer = WebsocketServer.New(&Config.WebsocketServer{
		InfoLoggerPath:    dashboardConfig.InfoLoggerPath,
		WarningLoggerPath: dashboardConfig.WarningLoggerPath,
		ErrorLoggerPath:   dashboardConfig.ErrorLoggerPath,
		Mailer:            dashboardConfig.Mailer,
		Pattern:           "/ws",
		ServerConfig: &Config.TcpServer{
			Port:        18251,
			TlsCertPath: dashboardConfig.ServerConfig.TlsCertPath,
			TlsKeyPath:  dashboardConfig.ServerConfig.TlsKeyPath,
			Blacklist:   dashboardConfig.ServerConfig.Blacklist,
			Whitelist:   dashboardConfig.ServerConfig.Whitelist,
		},
		ClientWatchdogTimeoutMs: 90000,
	}, app.GetWebsocketMessageHandlers(), app.OnConnectHandler, app.OnDisconnectHandler)

	app.systemgeServer = SystemgeServer.New(&Config.SystemgeServer{}, map[string]SystemgeServer.AsyncMessageHandler{}, map[string]SystemgeServer.SyncMessageHandler{})
	return app
}

func (app *App) Start() {
	go func() {
		err := app.httpServer.Start()
		if err != nil {
			panic(err)
		}
	}()
	go func() {
		err := app.websocketServer.Start()
		if err != nil {
			panic(err)
		}
	}()

	app.started = true
	_, filePath, _, _ := runtime.Caller(0)
	app.httpServer.AddRoute("/", HTTPServer.SendDirectory(filePath[:len(filePath)-len("lifecycle.go")]+"frontend"))

	if app.config.GoroutineUpdateIntervalMs > 0 {
		go app.goroutineUpdateRoutine()
	}
	if app.config.ServiceStatusUpdateIntervalMs > 0 {
		go app.serviceStatusUpdateRoutine()
	}
	if app.config.HeapUpdateIntervalMs > 0 {
		go app.heapUpdateRoutine()
	}

}

func (app *App) Stop() {
	app.started = false
	app.httpServer.Stop()
	app.websocketServer.Stop()
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
