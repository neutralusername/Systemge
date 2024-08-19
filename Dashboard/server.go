package Dashboard

/*
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

func NewDashboardServer(config *Config.DashboardServer) *DashboardServer {
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
	app := &DashboardServer{
		mutex:  sync.RWMutex{},
		config: config,

		infoLogger:    Tools.NewLogger("[Info: \"Dashboard\"]", config.InfoLoggerPath),
		warningLogger: Tools.NewLogger("[Warning: \"Dashboard\"]", config.WarningLoggerPath),
		errorLogger:   Tools.NewLogger("[Error: \"Dashboard\"]", config.ErrorLoggerPath),
		mailer:        Tools.NewMailer(config.MailerConfig),
	}
	app.httpServer = HTTPServer.New(config.HTTPServerConfig, nil)
	_, callerPath, _, _ := runtime.Caller(0)
	frontendPath := callerPath[:len(callerPath)-len("app.go")] + "frontend/"
	Helpers.CreateFile(frontendPath+"configs.js", "export const WS_PORT = "+Helpers.Uint16ToString(config.WebsocketServerConfig.TcpListenerConfig.Port)+";export const WS_PATTERN = \""+config.WebsocketServerConfig.Pattern+"\";")
	app.httpServer.AddRoute("/", HTTPServer.SendDirectory(frontendPath))

	app.websocketServer = WebsocketServer.New(config.WebsocketServerConfig, app.GetWebsocketMessageHandlers(), app.OnConnectHandler, nil)
	app.systemgeServer = SystemgeServer.New(config.SystemgeServerConfig, app.onSystemgeConnectHandler, nil, nil)

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
	app.registerModuleHttpHandlers(client)
	app.mutex.Lock()
	app.clients[client.Name] = client
	app.mutex.Unlock()
	return nil
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

func (app *DashboardServer) statusUpdateRoutine() {
	for !app.closed {
		app.mutex.RLock()
		for _, client := range app.clients {

		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.StatusUpdateIntervalMs) * time.Millisecond)
	}
}

func (app *DashboardServer) metricsUpdateRoutine() {
	for !app.closed {
		app.mutex.RLock()
		for _, client := range app.clients {

		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.MetricsUpdateIntervalMs) * time.Millisecond)
	}
}

func (app *DashboardServer) goroutineUpdateRoutine() {
	for !app.closed {
		goroutineCount := runtime.NumGoroutine()
		go app.websocketServer.Broadcast(Message.NewAsync("goroutineCount", strconv.Itoa(goroutineCount)))
		if infoLogger := app.infoLogger; infoLogger != nil {
			infoLogger.Log("goroutine update routine: \"" + strconv.Itoa(goroutineCount) + "\"")
		}
		time.Sleep(time.Duration(app.config.GoroutineUpdateIntervalMs) * time.Millisecond)
	}
}

func (app *DashboardServer) heapUpdateRoutine() {
	for !app.closed {
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
*/
