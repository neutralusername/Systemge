package DashboardClient

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
	"github.com/neutralusername/Systemge/TcpSystemgeConnection"
)

type Client struct {
	name                              string
	config                            *Config.DashboardClient
	dashboardServerSystemgeConnection SystemgeConnection.SystemgeConnection
	messageHandler                    SystemgeMessageHandler.MessageHandler
	introductionHandler               func() (string, error)
	messageHandlerStopChannel         chan<- bool

	status int
	mutex  sync.Mutex
}

func New(name string, config *Config.DashboardClient, messageHandler SystemgeMessageHandler.MessageHandler, introductionHandler func() (string, error)) *Client {
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
	if messageHandler == nil {
		panic("dashboardMessageHandler is nil")
	}

	app := &Client{
		name:                name,
		config:              config,
		messageHandler:      messageHandler,
		introductionHandler: introductionHandler,
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
	if message.GetTopic() != DashboardHelpers.TOPIC_INTRODUCTION {
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

	stopChannel, _ := SystemgeMessageHandler.StartMessageHandlingLoop_Sequentially(connection, app.messageHandler)
	app.messageHandlerStopChannel = stopChannel
	app.status = Status.STARTED
	return nil
}

func (app *Client) Stop() error {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.status == Status.STOPPED {
		return Error.New("Already stopped", nil)
	}
	close(app.messageHandlerStopChannel)
	app.dashboardServerSystemgeConnection.Close()
	app.dashboardServerSystemgeConnection = nil
	app.status = Status.STOPPED
	return nil
}
