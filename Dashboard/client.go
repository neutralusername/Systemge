package Dashboard

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
	"github.com/neutralusername/Systemge/SystemgeReceiver"
)

type Client struct {
	Name          string            `json:"name"`
	Commands      []string          `json:"commands"`
	Metrics       map[string]uint64 `json:"metrics"`
	HasStatusFunc bool              `json:"hasStatusFunc"`
	HasStartFunc  bool              `json:"hasStartFunc"`
	HasStopFunc   bool              `json:"hasStopFunc"`
}

func unmarshalClient(data string) (*Client, error) {
	var client Client
	err := json.Unmarshal([]byte(data), &client)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

type DashboardClient struct {
	config             *Config.DashboardClient
	systemgeConnection *SystemgeConnection.SystemgeConnection

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

	connection, err := SystemgeConnection.EstablishConnection(app.config.ConnectionConfig, app.config.EndpointConfig, app.config.Name, 0)
	if err != nil {
		panic(err)
	}
	app.systemgeConnection = connection
	var dashboardClientMessageHandlers = SystemgeMessageHandler.New(nil, map[string]func(*Message.Message) (string, error){
		Message.TOPIC_GET_INTRODUCTION: app.GetIntroductionHandler,
		Message.TOPIC_GET_STATUS:       app.GetStatusHandler,
		Message.TOPIC_GET_METRICS:      app.GetMetricsHandler,
		Message.TOPIC_START:            app.StartHandler,
		Message.TOPIC_STOP:             app.StopHandler,
		Message.TOPIC_EXECUTE_COMMAND:  app.ExecuteCommandHandler,
	})
	SystemgeReceiver.New(connection, app.config.ReceiverConfig, dashboardClientMessageHandlers)
	return app
}

func (app *DashboardClient) Close() {
	app.systemgeConnection.Close()
}

func (app *DashboardClient) GetIntroductionHandler(message *Message.Message) (string, error) {
	commands := []string{}
	for command := range app.commands {
		commands = append(commands, command)
	}
	return Helpers.JsonMarshal(&Client{
		Name:          app.config.Name,
		Commands:      commands,
		Metrics:       app.getMetricsFunc(),
		HasStatusFunc: app.getStatusFunc != nil,
		HasStartFunc:  app.startFunc != nil,
		HasStopFunc:   app.stopFunc != nil,
	}), nil
}

func (app *DashboardClient) GetStatusHandler(message *Message.Message) (string, error) {
	if app.getStatusFunc == nil {
		return "", Error.New("No status available", nil)
	}
	return Helpers.IntToString(app.getStatusFunc()), nil
}

func (app *DashboardClient) GetMetricsHandler(message *Message.Message) (string, error) {
	if app.getMetricsFunc == nil {
		return "", Error.New("No metrics available", nil)
	}
	return Helpers.JsonMarshal(app.getMetricsFunc()), nil
}

func (app *DashboardClient) StartHandler(message *Message.Message) (string, error) {
	if app.startFunc == nil {
		return "", Error.New("No start function available", nil)
	}
	err := app.startFunc()
	if err != nil {
		return "", err
	}
	return "", nil
}

func (app *DashboardClient) StopHandler(message *Message.Message) (string, error) {
	if app.stopFunc == nil {
		return "", Error.New("No stop function available", nil)
	}
	err := app.stopFunc()
	if err != nil {
		return "", err
	}
	return "", nil
}

func (app *DashboardClient) ExecuteCommandHandler(message *Message.Message) (string, error) {
	if app.commands == nil {
		return "", nil
	}
	command := message.GetPayload()
	commandFunc, ok := app.commands[command]
	if !ok {
		return "", Error.New("Command not found", nil)
	}
	return "", commandFunc(nil)
}
