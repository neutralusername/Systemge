package Dashboard

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
)

type DashboardClient struct {
	config             *Config.DashboardClient
	systemgeConnection *SystemgeConnection.SystemgeConnection

	startFunc      func() error
	stopFunc       func() error
	getMetricsFunc func() Metrics
	getStatusFunc  func() int
	commands       CommandHandlers
}

func NewClient(config *Config.DashboardClient, startFunc func() error, stopFunc func() error, getMetricsFunc func() Metrics, getStatusFunc func() int, commands CommandHandlers) *DashboardClient {
	if config == nil {
		panic("config is nil")
	}
	if config.Name == "" {
		panic("config.Name is empty")
	}
	if config.ConnectionConfig == nil {
		panic("config.ConnectionConfig is nil")
	}
	if config.ConnectionConfig.TcpBufferBytes == 0 {
		config.ConnectionConfig.TcpBufferBytes = 1024 * 4
	}
	if config.EndpointConfig == nil {
		panic("config.EndpointConfig is nil")
	}
	app := &DashboardClient{
		config:         config,
		startFunc:      startFunc,
		stopFunc:       stopFunc,
		getMetricsFunc: getMetricsFunc,
		getStatusFunc:  getStatusFunc,
		commands:       commands,
	}
	var dashboardClientMessageHandlers = SystemgeMessageHandler.New(nil, map[string]func(*Message.Message) (string, error){
		Message.TOPIC_GET_INTRODUCTION: app.getIntroductionHandler,
		Message.TOPIC_GET_STATUS:       app.getStatusHandler,
		Message.TOPIC_GET_METRICS:      app.getMetricsHandler,
		Message.TOPIC_START:            app.startHandler,
		Message.TOPIC_STOP:             app.stopHandler,
		Message.TOPIC_EXECUTE_COMMAND:  app.executeCommandHandler,
	})
	connection, err := SystemgeConnection.EstablishConnection(app.config.ConnectionConfig, app.config.EndpointConfig, app.config.Name, 0, dashboardClientMessageHandlers)
	if err != nil {
		panic(err)
	}
	app.systemgeConnection = connection
	if config.ConnectionConfig.ProcessSequentially {
		app.systemgeConnection.StartProcessingLoopSequentially()
	} else {
		app.systemgeConnection.StartProcessingLoopConcurrently()
	}
	return app
}

func (app *DashboardClient) Close() {
	app.systemgeConnection.Close()
}

func (app *DashboardClient) getIntroductionHandler(message *Message.Message) (string, error) {
	commands := make(map[string]bool)
	for command := range app.commands {
		commands[command] = true
	}
	status := Status.NON_EXISTENT
	if app.getStatusFunc != nil {
		status = app.getStatusFunc()
	}
	metrics := make(Metrics)
	if app.getMetricsFunc != nil {
		metrics = app.getMetricsFunc()
	}
	return Helpers.JsonMarshal(&client{
		Name:           app.config.Name,
		Status:         status,
		Metrics:        metrics,
		Commands:       commands,
		HasStatusFunc:  app.getStatusFunc != nil,
		HasStartFunc:   app.startFunc != nil,
		HasStopFunc:    app.stopFunc != nil,
		HasMetricsFunc: app.getMetricsFunc != nil,
	}), nil
}

func (app *DashboardClient) getStatusHandler(message *Message.Message) (string, error) {
	if app.getStatusFunc == nil {
		return "", Error.New("No status available", nil)
	}
	return Helpers.IntToString(app.getStatusFunc()), nil
}

func (app *DashboardClient) getMetricsHandler(message *Message.Message) (string, error) {
	if app.getMetricsFunc == nil {
		return "", Error.New("No metrics available", nil)
	}
	metrics := metrics{
		Metrics: app.getMetricsFunc(),
		Name:    app.config.Name,
	}
	return Helpers.JsonMarshal(metrics), nil
}

func (app *DashboardClient) startHandler(message *Message.Message) (string, error) {
	if app.startFunc == nil {
		return "", Error.New("No start function available", nil)
	}
	err := app.startFunc()
	if err != nil {
		return "", err
	}
	return Helpers.IntToString(app.getStatusFunc()), nil
}

func (app *DashboardClient) stopHandler(message *Message.Message) (string, error) {
	if app.stopFunc == nil {
		return "", Error.New("No stop function available", nil)
	}
	err := app.stopFunc()
	if err != nil {
		return "", err
	}
	return Helpers.IntToString(app.getStatusFunc()), nil
}

func (app *DashboardClient) executeCommandHandler(message *Message.Message) (string, error) {
	if app.commands == nil {
		return "", nil
	}
	command, err := unmarshalCommand(message.GetPayload())
	if err != nil {
		return "", err
	}
	commandFunc, ok := app.commands[command.Command]
	if !ok {
		return "", Error.New("Command not found", nil)
	}
	return commandFunc(command.Args)
}
