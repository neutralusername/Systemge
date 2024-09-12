package DashboardClientSystemgeConnection

import (
	"sync"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/TcpSystemgeConnection"
)

type Client struct {
	name                     string
	config                   *Config.DashboardClient
	serverSystemgeConnection SystemgeConnection.SystemgeConnection

	clientSystemgeConnection SystemgeConnection.SystemgeConnection

	commands Commands.Handlers

	messageHandler *SystemgeConnection.TopicExclusiveMessageHandler

	status int
	mutex  sync.Mutex
}

func New(name string, config *Config.DashboardClient, systemgeConnection SystemgeConnection.SystemgeConnection, messageHandler SystemgeConnection.MessageHandler, commands Commands.Handlers) *Client {
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
	if systemgeConnection == nil {
		panic("customService is nil")
	}
	app := &Client{
		name:                     name,
		config:                   config,
		clientSystemgeConnection: systemgeConnection,
		commands:                 commands,
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
	app.serverSystemgeConnection = connection

	message, err := connection.GetNextMessage()
	if err != nil {
		connection.Close()
		return Error.New("Failed to get introduction message", err)
	}
	if message.GetTopic() != Message.TOPIC_INTRODUCTION {
		connection.Close()
		return Error.New("Expected introduction message", nil)
	}
	response, err := app.introductionHandler()
	if err != nil {
		connection.Close()
		return Error.New("Failed to handle introduction message", err)
	}
	err = connection.SyncResponse(message, true, response)
	if err != nil {
		connection.Close()
		return Error.New("Failed to send introduction response", err)
	}

	app.messageHandler = SystemgeConnection.NewTopicExclusiveMessageHandler(
		nil,
		SystemgeConnection.SyncMessageHandlers{
			Message.TOPIC_GET_STATUS:       app.getStatusHandler,
			Message.TOPIC_GET_METRICS:      app.getMetricsHandler,
			Message.TOPIC_CLOSE_CONNECTION: app.closeConnectionHandler,
			Message.TOPIC_EXECUTE_COMMAND:  app.executeCommandHandler,
			Message.TOPIC_START_PROCESSINGLOOP_SEQUENTIALLY: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				return "", Error.New("Cannot start processing loop sequentially", nil)
			},
			Message.TOPIC_START_PROCESSINGLOOP_CONCURRENTLY: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				return "", Error.New("Cannot start processing loop concurrently", nil)
			},
			Message.TOPIC_STOP_PROCESSINGLOOP: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				return "", Error.New("Cannot stop processing loop", nil)
			},
			Message.PROCESS_NEXT_MESSAGE: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				return "", Error.New("Cannot handle next message", nil)
			},
			Message.UNPROCESSED_MESSAGES_COUNT: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				return "", Error.New("Cannot get unprocessed messages count", nil)
			},
			Message.SYNC_REQUEST: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				return "", Error.New("Cannot sync request", nil)
			},
			Message.ASYNC_MESSAGE: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				return "", Error.New("Cannot handle async message", nil)
			},
		},
		nil, nil,
		100,
	)
	app.serverSystemgeConnection.StartProcessingLoopSequentially(app.messageHandler)
	app.status = Status.STARTED
	return nil
}

func (app *Client) introductionHandler() (string, error) {
	return string(DashboardHelpers.NewIntroduction(
		DashboardHelpers.NewCustomServiceClient(
			app.name,
			app.commands.GetKeys(),
			app.clientSystemgeConnection.GetStatus(),
			app.clientSystemgeConnection.GetMetrics(),
		),
		DashboardHelpers.CLIENT_CUSTOM_SERVICE,
	).Marshal()), nil
}

func (app *Client) closeConnectionHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	err := app.clientSystemgeConnection.Close()
	if err != nil {
		return "", err
	}
	return "", nil
}

func (app *Client) getStatusHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	return Helpers.IntToString(app.clientSystemgeConnection.GetStatus()), nil
}

func (app *Client) getMetricsHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	return DashboardHelpers.NewMetrics(app.name, app.clientSystemgeConnection.GetMetrics()).Marshal(), nil
}

func (app *Client) executeCommandHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	command, err := DashboardHelpers.UnmarshalCommand(message.GetPayload())
	if err != nil {
		return "", err
	}
	return app.commands.Execute(command.Command, command.Args)
}
