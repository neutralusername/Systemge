package DashboardClientCommands

import (
	"sync"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/TcpSystemgeConnection"
)

type Client struct {
	name               string
	config             *Config.DashboardClient
	systemgeConnection SystemgeConnection.SystemgeConnection

	messageHandler *SystemgeConnection.TopicExclusiveMessageHandler

	commands Commands.Handlers

	status int
	mutex  sync.Mutex
}

func New(name string, config *Config.DashboardClient, commands Commands.Handlers) *Client {
	if config == nil {
		panic("config is nil")
	}
	if name == "" {
		panic("config.Name is empty")
	}
	if config.TcpSystemgeConnectionConfig == nil {
		panic("config.ConnectionConfig is nil")
	}
	if config.TcpClientConfig == nil {
		panic("config.EndpointConfig is nil")
	}
	app := &Client{
		name:     name,
		config:   config,
		commands: commands,
	}
	return app
}

func (app *Client) Start() error {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.status == Status.STARTED {
		return Error.New("Already started", nil)
	}
	connection, err := TcpSystemgeConnection.EstablishConnection(app.config.TcpSystemgeConnectionConfig, app.config.TcpClientConfig, app.name, app.config.MaxServerNameLength)
	if err != nil {
		return Error.New("Failed to establish connection", err)
	}
	app.systemgeConnection = connection
	app.messageHandler = SystemgeConnection.NewTopicExclusiveMessageHandler(
		nil,
		SystemgeConnection.SyncMessageHandlers{
			Message.TOPIC_INTRODUCTION:    app.introductionHandler,
			Message.TOPIC_EXECUTE_COMMAND: app.executeCommandHandler,
		},
		nil, nil,
		100,
	)
	app.systemgeConnection.StartProcessingLoopSequentially(app.messageHandler)
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
	app.messageHandler.Close()
	app.messageHandler = nil
	app.systemgeConnection.Close()
	app.systemgeConnection = nil
	app.status = Status.STOPPED
	return nil
}

func (app *Client) introductionHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	return string(DashboardHelpers.NewIntroduction(
		DashboardHelpers.NewCommandClient(
			app.name,
			app.commands.GetKeys(),
		),
		DashboardHelpers.CLIENT_COMMAND,
	).Marshal()), nil
}

func (app *Client) executeCommandHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	command, err := DashboardHelpers.UnmarshalCommand(message.GetPayload())
	if err != nil {
		return "", err
	}
	return app.commands.Execute(command.Command, command.Args)
}
