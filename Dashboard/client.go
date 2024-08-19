package Dashboard

/*
type DashboardClient struct {
	config             *Config.DashboardClient
	systemgeConnection *SystemgeConnection.SystemgeConnection

	name           string
	startFunc      func() error
	stopFunc       func() error
	getMetricsFunc func() map[string]uint64
	getStatusFunc  func() int
	commands       map[string]func(args []string) error
}

func NewDashboardClient(config *Config.DashboardClient, startFunc func() error, stopFunc func() error, getMetricsFunc func() map[string]uint64, getStatusFunc func() int, commands map[string]func(args []string) error) *DashboardClient {
	if config == nil {
		panic("config is nil")
	}
	if config.ConnectionConfig == nil {
		panic("config.ConnectionConfig is nil")
	}
	if config.EndpointConfig == nil {
		panic("config.EndpointConfig is nil")
	}
	if config.ReceiverConfig == nil {
		panic("config.ReceiverConfig is nil")
	}
	app := &DashboardClient{
		config:         config,
		startFunc:      startFunc,
		stopFunc:       stopFunc,
		getMetricsFunc: getMetricsFunc,
		getStatusFunc:  getStatusFunc,
		commands:       commands,
	}

	connection, err := SystemgeConnection.EstablishConnection(app.config.ConnectionConfig, app.config.EndpointConfig, app.name, 0)
	if err != nil {
		panic(err)
	}
	app.systemgeConnection = connection
	var dashboardClientMessageHandlers = SystemgeMessageHandler.New(nil, map[string]func(*Message.Message) (string, error){
		Message.TOPIC_GET_SERVICE_STATUS: app.GetServiceStatusHandler,
		Message.TOPIC_GET_METRICS:        app.GetMetricsHandler,
		Message.TOPIC_START_SERVICE:      app.StartServiceHandler,
		Message.TOPIC_STOP_SERVICE:       app.StopServiceHandler,
		Message.TOPIC_EXECUTE_COMMAND:    app.ExecuteCommandHandler,
	})
	SystemgeReceiver.New(connection, app.config.ReceiverConfig, dashboardClientMessageHandlers)
	return app
}

func (app *DashboardClient) Close() {
	app.systemgeConnection.Close()
}

func (app *DashboardClient) GetServiceStatusHandler(message *Message.Message) (string, error) {
	if app.getStatusFunc == nil {
		return "", nil
	}
	return Helpers.IntToString(app.getStatusFunc()), nil
}

func (app *DashboardClient) GetMetricsHandler(message *Message.Message) (string, error) {

}

func (app *DashboardClient) StartServiceHandler(message *Message.Message) (string, error) {

}

func (app *DashboardClient) StopServiceHandler(message *Message.Message) (string, error) {

}

func (app *DashboardClient) ExecuteCommandHandler(message *Message.Message) (string, error) {

}
*/
