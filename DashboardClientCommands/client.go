package DashboardClientCommands

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

type CustomClient struct {
	name               string
	config             *Config.DashboardClient
	systemgeConnection SystemgeConnection.SystemgeConnection

	commands Commands.Handlers

	status int
	mutex  sync.Mutex
}

func New(name string, config *Config.DashboardClient, commands Commands.Handlers) *CustomClient {
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
	app := &CustomClient{
		name:     name,
		config:   config,
		commands: commands,
	}
	return app
}

func (app *CustomClient) Start() error {
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
			Message.TOPIC_EXECUTE_COMMAND:  app.executeCommandHandler,
		}, nil, nil),
	)

	app.status = Status.STARTED
	return nil
}

func (app *CustomClient) Stop() error {
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

func (app *CustomClient) getIntroductionHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	commands := []string{}
	for command := range app.commands {
		commands = append(commands, command)
	}
	return Helpers.JsonMarshal(&DashboardUtilities.Introduction{
		Client: &DashboardUtilities.CommandClient{
			Name:     app.name,
			Commands: commands,
		},
		ClientType: DashboardUtilities.CLIENT_COMMAND,
	}), nil
}

func (app *CustomClient) executeCommandHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
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
