package DashboardClient

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/TcpSystemgeConnection"
)

type Client struct {
	name                              string
	config                            *Config.DashboardClient
	dashboardServerSystemgeConnection SystemgeConnection.SystemgeConnection
	messageHandler                    *SystemgeConnection.SequentialMessageHandler
	asyncMessageHandlerFuncs          SystemgeConnection.AsyncMessageHandlers
	syncMessageHandlerFuncs           SystemgeConnection.SyncMessageHandlers
	introductionHandler               func() (string, error)

	status int
	mutex  sync.Mutex
}

func New(name string, config *Config.DashboardClient, asyncMessageHandlerFuncs SystemgeConnection.AsyncMessageHandlers, syncMessageHandlerFuncs SystemgeConnection.SyncMessageHandlers, introductionHandler func() (string, error)) *Client {
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
		panic("config.TcpClientConfig is nil")
	}

	app := &Client{
		name:                     name,
		config:                   config,
		asyncMessageHandlerFuncs: asyncMessageHandlerFuncs,
		syncMessageHandlerFuncs:  syncMessageHandlerFuncs,
		introductionHandler:      introductionHandler,
	}
	return app
}

func (app *Client) Start() error {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.status == Status.STARTED {
		return Event.New("Already started", nil)
	}
	connection, err := TcpSystemgeConnection.EstablishConnection(app.config.TcpSystemgeConnectionConfig, app.config.TcpClientConfig, app.name, app.config.MaxServerNameLength)
	if err != nil {
		return Event.New("Failed to establish connection", err)
	}
	app.dashboardServerSystemgeConnection = connection

	message, err := connection.GetNextMessage()
	if err != nil {
		connection.Close()
		return Event.New("Failed to get introduction message", err)
	}
	if message.GetTopic() != DashboardHelpers.TOPIC_INTRODUCTION {
		connection.Close()
		return Event.New("Expected introduction message", nil)
	}
	response, err := app.introductionHandler()
	if err != nil {
		connection.Close()
		return Event.New("Failed to handle introduction message", err)
	}
	err = connection.SyncResponse(message, true, response)
	if err != nil {
		connection.Close()
		return Event.New("Failed to send introduction response", err)
	}

	app.messageHandler = SystemgeConnection.NewSequentialMessageHandler(
		app.asyncMessageHandlerFuncs,
		app.syncMessageHandlerFuncs,
		nil, nil,
		app.config.MessageHandlerQueueSize,
	)

	err = connection.StartMessageHandlingLoop_Sequentially(app.messageHandler)
	if err != nil {
		connection.Close()
		return Event.New("Failed to start message handling loop", err)
	}
	app.status = Status.STARTED
	return nil
}

func (app *Client) Stop() error {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.status == Status.STOPPED {
		return Event.New("Already stopped", nil)
	}
	app.dashboardServerSystemgeConnection.Close()
	app.dashboardServerSystemgeConnection = nil
	app.messageHandler.Close()
	app.messageHandler = nil
	app.status = Status.STOPPED
	return nil
}
