package Dashboard

import (
	"sync"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/TcpSystemgeConnection"
)

type Client struct {
	name               string
	config             *Config.DashboardClient
	systemgeConnection SystemgeConnection.SystemgeConnection

	startFunc      func() error
	stopFunc       func() error
	getMetricsFunc func() map[string]uint64
	getStatusFunc  func() int
	commands       Commands.Handlers

	status int
	mutex  sync.Mutex
}

func NewClient(name string, config *Config.DashboardClient, startFunc func() error, stopFunc func() error, getMetricsFunc func() map[string]uint64, getStatusFunc func() int, commands Commands.Handlers) *Client {
	if config == nil {
		panic("config is nil")
	}
	if name == "" {
		panic("config.Name is empty")
	}
	if config.ConnectionConfig == nil {
		panic("config.ConnectionConfig is nil")
	}
	if config.ClientConfig == nil {
		panic("config.EndpointConfig is nil")
	}
	app := &Client{
		name:           name,
		config:         config,
		startFunc:      startFunc,
		stopFunc:       stopFunc,
		getMetricsFunc: getMetricsFunc,
		getStatusFunc:  getStatusFunc,
		commands:       commands,
	}
	return app
}

func (app *Client) Start() error {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.status == Status.STARTED {
		return Error.New("Already started", nil)
	}
	connection, err := TcpSystemgeConnection.EstablishConnection(app.config.ConnectionConfig, app.config.ClientConfig, app.name, app.config.MaxServerNameLength)
	if err != nil {
		return Error.New("Failed to establish connection", err)
	}
	app.systemgeConnection = connection
	app.systemgeConnection.StartProcessingLoopSequentially(
		SystemgeConnection.NewConcurrentMessageHandler(nil, SystemgeConnection.SyncMessageHandlers{
			Message.TOPIC_GET_INTRODUCTION: app.getIntroductionHandler,
			Message.TOPIC_GET_STATUS:       app.getStatusHandler,
			Message.TOPIC_GET_METRICS:      app.getMetricsHandler,
			Message.TOPIC_START:            app.startHandler,
			Message.TOPIC_STOP:             app.stopHandler,
			Message.TOPIC_EXECUTE_COMMAND:  app.executeCommandHandler,
		}, nil, nil),
	)
	app.status = Status.STARTED
	return nil
}

func (app *Client) Stop() error {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.status == Status.STOPPED {
		return Error.New("Already stopped", nil)
	}
	app.systemgeConnection.StopProcessingLoop()
	app.systemgeConnection.Close()
	app.systemgeConnection = nil
	app.status = Status.STOPPED
	return nil
}

func (app *Client) getIntroductionHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	commands := make(map[string]bool)
	for command := range app.commands {
		commands[command] = true
	}
	status := Status.NON_EXISTENT
	if app.getStatusFunc != nil {
		status = app.getStatusFunc()
	}
	metrics := make(map[string]uint64)
	if app.getMetricsFunc != nil {
		metrics = app.getMetricsFunc()
	}
	return Helpers.JsonMarshal(&client{
		Name:           app.name,
		Status:         status,
		Metrics:        metrics,
		Commands:       commands,
		HasStatusFunc:  app.getStatusFunc != nil,
		HasStartFunc:   app.startFunc != nil,
		HasStopFunc:    app.stopFunc != nil,
		HasMetricsFunc: app.getMetricsFunc != nil,
	}), nil
}

func (app *Client) getStatusHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	if app.getStatusFunc == nil {
		return "", Error.New("No status available", nil)
	}
	return Helpers.IntToString(app.getStatusFunc()), nil
}

func (app *Client) getMetricsHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	if app.getMetricsFunc == nil {
		return "", Error.New("No metrics available", nil)
	}
	metrics := metrics{
		Metrics: app.getMetricsFunc(),
		Name:    app.name,
	}
	return Helpers.JsonMarshal(metrics), nil
}

func (app *Client) startHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	if app.startFunc == nil {
		return "", Error.New("No start function available", nil)
	}
	err := app.startFunc()
	if err != nil {
		return "", err
	}
	return Helpers.IntToString(app.getStatusFunc()), nil
}

func (app *Client) stopHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	if app.stopFunc == nil {
		return "", Error.New("No stop function available", nil)
	}
	err := app.stopFunc()
	if err != nil {
		return "", err
	}
	return Helpers.IntToString(app.getStatusFunc()), nil
}

func (app *Client) executeCommandHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
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
