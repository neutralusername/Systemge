package DashboardClient

import (
	"errors"
	"sync"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/DashboardHelpers"
	"github.com/neutralusername/systemge/Event"
	"github.com/neutralusername/systemge/SystemgeConnection"
	"github.com/neutralusername/systemge/TcpConnect"
	"github.com/neutralusername/systemge/status"
)

type Client struct {
	name                              string
	config                            *configs.DashboardClient
	dashboardServerSystemgeConnection SystemgeConnection.SystemgeConnection
	messageHandler                    *SystemgeConnection.SequentialMessageHandler
	asyncMessageHandlerFuncs          SystemgeConnection.AsyncMessageHandlers
	syncMessageHandlerFuncs           SystemgeConnection.SyncMessageHandlers
	introductionHandler               func() (string, error)

	eventHandler Event.Handler

	status int
	mutex  sync.Mutex
}

func New(name string, config *configs.DashboardClient, asyncMessageHandlerFuncs SystemgeConnection.AsyncMessageHandlers, syncMessageHandlerFuncs SystemgeConnection.SyncMessageHandlers, introductionHandler func() (string, error), eventHandler Event.Handler) (*Client, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if name == "" {
		return nil, errors.New("name is empty")
	}
	if config.TcpSystemgeConnectionConfig == nil {
		return nil, errors.New("config.TcpSystemgeConnectionConfig is nil")
	}
	if config.TcpClientConfig == nil {
		return nil, errors.New("config.TcpClientConfig is nil")
	}

	app := &Client{
		name:                     name,
		config:                   config,
		asyncMessageHandlerFuncs: asyncMessageHandlerFuncs,
		syncMessageHandlerFuncs:  syncMessageHandlerFuncs,
		introductionHandler:      introductionHandler,

		eventHandler: eventHandler,
	}
	return app, nil
}

func (app *Client) Start() error {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.status == status.Started {
		return errors.New("Already started")
	}
	connection, err := TcpConnect.EstablishConnection(app.config.TcpSystemgeConnectionConfig, app.config.TcpClientConfig, app.name, app.config.MaxServerNameLength, app.eventHandler)
	if err != nil {
		return err
	}
	app.dashboardServerSystemgeConnection = connection

	message, err := connection.RetrieveNextMessage()
	if err != nil {
		connection.Close()
		return err
	}
	if message.GetTopic() != DashboardHelpers.TOPIC_INTRODUCTION {
		connection.Close()
		return errors.New("Expected introduction message")
	}
	response, err := app.introductionHandler()
	if err != nil {
		connection.Close()
		return err
	}
	err = connection.SyncResponse(message, true, response)
	if err != nil {
		connection.Close()
		return err
	}

	app.messageHandler = SystemgeConnection.NewSequentialMessageHandler(
		app.asyncMessageHandlerFuncs,
		app.syncMessageHandlerFuncs,
		nil, nil,
		app.config.MessageHandlerQueueSize,
	)

	err = connection.StartMessageHandlingLoop(app.messageHandler, true)
	if err != nil {
		connection.Close()
		return err
	}
	app.status = status.Started
	return nil
}

func (app *Client) Stop() error {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.status == status.Stopped {
		return errors.New("Already stopped")
	}
	app.dashboardServerSystemgeConnection.Close()
	app.dashboardServerSystemgeConnection = nil
	app.messageHandler.Close()
	app.messageHandler = nil
	app.status = status.Stopped
	return nil
}
