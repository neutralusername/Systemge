package DashboardClientCustomService

import (
	"sync"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/DashboardUtilities"
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

	customService CustomService
	commands      Commands.Handlers

	status int
	mutex  sync.Mutex
}

type CustomService interface {
	Start() error
	Stop() error
	GetStatus() int
	GetMetrics() map[string]uint64
}

func New(name string, config *Config.DashboardClient, customService CustomService, commands Commands.Handlers) *Client {
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
	if customService == nil {
		panic("customService is nil")
	}
	app := &Client{
		name:          name,
		config:        config,
		customService: customService,
		commands:      commands,
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
		SystemgeConnection.NewConcurrentMessageHandler(
			nil,
			SystemgeConnection.SyncMessageHandlers{
				Message.TOPIC_INTRODUCTION:    app.introductionHandler,
				Message.TOPIC_GET_STATUS:      app.getStatusHandler,
				Message.TOPIC_GET_METRICS:     app.getMetricsHandler,
				Message.TOPIC_START:           app.startHandler,
				Message.TOPIC_STOP:            app.stopHandler,
				Message.TOPIC_EXECUTE_COMMAND: app.executeCommandHandler,
			},
			nil, nil,
		),
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

func (app *Client) introductionHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	return string(DashboardUtilities.NewClient(
		DashboardUtilities.NewCustomServiceClient(
			app.name,
			app.commands.GetKeys(),
			app.customService.GetStatus(),
			app.systemgeConnection.GetMetrics(),
		),
		DashboardUtilities.CLIENT_CUSTOM_SERVICE,
	).Marshal()), nil
}

func (app *Client) getStatusHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	return Helpers.IntToString(app.customService.GetStatus()), nil
}

func (app *Client) getMetricsHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	return DashboardUtilities.NewMetrics(app.name, app.customService.GetMetrics()).Marshal(), nil
}

func (app *Client) startHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	err := app.customService.Start()
	if err != nil {
		return "", err
	}
	return Helpers.IntToString(app.customService.GetStatus()), nil
}

func (app *Client) stopHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	err := app.customService.Start()
	if err != nil {
		return "", err
	}
	return Helpers.IntToString(app.customService.GetStatus()), nil
}

func (app *Client) executeCommandHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	if app.commands == nil {
		return "", nil
	}
	command, err := DashboardUtilities.UnmarshalCommand(message.GetPayload())
	if err != nil {
		return "", err
	}
	commandFunc, ok := app.commands[command.Command]
	if !ok {
		return "", Error.New("Command not found", nil)
	}
	return commandFunc(command.Args)
}
