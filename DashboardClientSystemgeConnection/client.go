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
	name                              string
	config                            *Config.DashboardClient
	dashboardServerSystemgeConnection SystemgeConnection.SystemgeConnection
	dashboardClientMessageHandler     *SystemgeConnection.TopicExclusiveMessageHandler

	systemgeConnection SystemgeConnection.SystemgeConnection
	messageHandler     SystemgeConnection.MessageHandler
	commands           Commands.Handlers

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
		name:               name,
		config:             config,
		systemgeConnection: systemgeConnection,
		commands:           commands,
		messageHandler:     messageHandler,
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
	app.dashboardServerSystemgeConnection = connection

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

	app.dashboardClientMessageHandler = SystemgeConnection.NewTopicExclusiveMessageHandler(
		nil,
		SystemgeConnection.SyncMessageHandlers{
			Message.TOPIC_GET_STATUS:                        app.getStatusHandler,
			Message.TOPIC_GET_METRICS:                       app.getMetricsHandler,
			Message.TOPIC_CLOSE_CONNECTION:                  app.closeConnectionHandler,
			Message.TOPIC_EXECUTE_COMMAND:                   app.executeCommandHandler,
			Message.TOPIC_START_PROCESSINGLOOP_SEQUENTIALLY: app.startProcessingLoopSequentially,
			Message.TOPIC_START_PROCESSINGLOOP_CONCURRENTLY: app.startProcessingLoopConcurrently,
			Message.TOPIC_STOP_PROCESSINGLOOP:               app.stopProcessingLoop,
			Message.TOPIC_IS_PROCESSING_LOOP_RUNNING:        app.isProcessingLoopRunningHandler,
			Message.TOPIC_PROCESS_NEXT_MESSAGE:              app.processNextMessage,
			Message.TOPIC_UNPROCESSED_MESSAGES_COUNT:        app.unprocessedMessagesCount,
			Message.TOPIC_SYNC_REQUEST:                      app.syncRequestHandler,
			Message.TOPIC_ASYNC_MESSAGE:                     app.asyncMessageHandler,
		},
		nil, nil,
		100,
	)
	app.dashboardServerSystemgeConnection.StartProcessingLoopSequentially(app.dashboardClientMessageHandler)
	app.status = Status.STARTED
	return nil
}

func (app *Client) Stop() error {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.status == Status.STOPPED {
		return Error.New("Already stopped", nil)
	}
	app.dashboardClientMessageHandler.Close()
	app.dashboardClientMessageHandler = nil
	app.dashboardServerSystemgeConnection.StopProcessingLoop()
	app.dashboardServerSystemgeConnection.Close()
	app.dashboardServerSystemgeConnection = nil
	app.status = Status.STOPPED
	return nil
}

func (app *Client) SetMessageHandler(messageHandler SystemgeConnection.MessageHandler) {
	app.messageHandler = messageHandler
}

func (app *Client) introductionHandler() (string, error) {
	return string(DashboardHelpers.NewIntroduction(
		DashboardHelpers.NewSystemgeConnectionClient(
			app.name,
			app.commands.GetKeys(),
			app.systemgeConnection.GetStatus(),
			app.systemgeConnection.GetMetrics(),
			app.systemgeConnection.UnprocessedMessagesCount(),
		),
	).Marshal()), nil
}

func (app *Client) closeConnectionHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	err := app.systemgeConnection.Close()
	if err != nil {
		return "", err
	}
	return Helpers.IntToString(Status.STOPPED), nil
}

func (app *Client) getStatusHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	return Helpers.IntToString(app.systemgeConnection.GetStatus()), nil
}

func (app *Client) getMetricsHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	return DashboardHelpers.NewMetrics(app.name, app.systemgeConnection.GetMetrics()).Marshal(), nil
}

func (app *Client) executeCommandHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	command, err := DashboardHelpers.UnmarshalCommand(message.GetPayload())
	if err != nil {
		return "", err
	}
	return app.commands.Execute(command.Command, command.Args)
}

func (app *Client) startProcessingLoopSequentially(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	err := app.systemgeConnection.StartProcessingLoopSequentially(app.messageHandler)
	if err != nil {
		return "", err
	}
	return "", nil
}

func (app *Client) startProcessingLoopConcurrently(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	err := app.systemgeConnection.StartProcessingLoopConcurrently(app.messageHandler)
	if err != nil {
		return "", err
	}
	return "", nil
}

func (app *Client) stopProcessingLoop(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	err := app.systemgeConnection.StopProcessingLoop()
	if err != nil {
		return "", err
	}
	return "", nil
}

func (app *Client) isProcessingLoopRunningHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	return Helpers.BoolToString(app.systemgeConnection.IsProcessingLoopRunning()), nil
}

func (app *Client) processNextMessage(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	message, err := app.systemgeConnection.GetNextMessage()
	if err != nil {
		return "", Error.New("Failed to get next message", err)
	}
	if message.GetSyncToken() == "" {
		err := app.messageHandler.HandleAsyncMessage(app.systemgeConnection, message)
		if err != nil {
			return "", Error.New("Failed to handle async message with topic \""+message.GetTopic()+"\" and payload \""+message.GetPayload()+"\"", err)
		}
		return "Handled async message with topic \"" + message.GetTopic() + "\" and payload \"" + message.GetPayload() + "\"", nil
	}
	if responsePayload, err := app.messageHandler.HandleSyncRequest(app.systemgeConnection, message); err != nil {
		if err := app.systemgeConnection.SyncResponse(message, false, err.Error()); err != nil {
			return "", Error.New("Failed to handle sync request with topic \""+message.GetTopic()+"\" and payload \""+message.GetPayload()+"\" and failed to send failure response \""+err.Error()+"\"", err)
		}
		return "Failed to handle sync request with topic \"" + message.GetTopic() + "\" and payload \"" + message.GetPayload() + "\" and sent failure response \"" + err.Error() + "\"", nil
	} else {
		if err := app.systemgeConnection.SyncResponse(message, true, responsePayload); err != nil {
			return "", Error.New("Handled sync request with topic \""+message.GetTopic()+"\" and payload \""+message.GetPayload()+"\" and failed to send success response \""+responsePayload+"\"", err)
		}
		return "Handled sync request with topic \"" + message.GetTopic() + "\" and payload \"" + message.GetPayload() + "\" and sent success response \"" + responsePayload + "\"", nil
	}
}

func (app *Client) unprocessedMessagesCount(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	return Helpers.Uint32ToString(app.systemgeConnection.UnprocessedMessagesCount()), nil
}

func (app *Client) syncRequestHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	message, err := Message.Deserialize([]byte(message.GetPayload()), connection.GetName())
	if err != nil {
		return "", Error.New("Failed to deserialize message", err)
	}
	response, err := app.systemgeConnection.SyncRequestBlocking(message.GetTopic(), message.GetPayload())
	if err != nil {
		return "", Error.New("Failed to complete sync request", err)
	}
	return string(response.Serialize()), nil
}

func (app *Client) asyncMessageHandler(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	message, err := Message.Deserialize([]byte(message.GetPayload()), connection.GetName())
	if err != nil {
		return "", Error.New("Failed to deserialize message", err)
	}
	err = app.systemgeConnection.AsyncMessage(message.GetTopic(), message.GetPayload())
	if err != nil {
		return "", Error.New("Failed to handle async message", err)
	}
	return "", nil
}
