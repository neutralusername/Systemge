package MessageBrokerClient

import (
	"Systemge/Application"
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"Systemge/WebsocketServer"
	"sync"
)

type Client struct {
	name                         string
	logger                       *Utilities.Logger
	topicResolutionServerAddress string
	randomizer                   *Utilities.Randomizer

	websocketServer *WebsocketServer.Server
	application     Application.Application

	serverConnections          map[string]*serverConnection     // address -> serverConnection
	topicResolutions           map[string]*serverConnection     // topic -> serverConnection
	messagesWaitingForResponse map[string]chan *Message.Message // syncKey -> responseChannel
	mapOperationMutex          sync.Mutex

	// handleServerMessagesConcurrently is a flag that determines whether the client will handle messages concurrently
	// If this is set to true, the client will handle messages concurrently, otherwise it will handle messages sequentially
	// concurrently handling messages can be useful to improve performance but it will often require additional work for the application to be able to handle concurrency
	handleServerMessagesConcurrently      bool
	handleServerMessagesConcurrentlyMutex sync.Mutex

	stopChannel chan bool
	isStarted   bool

	stateMutex sync.Mutex
}

func New(name, topicResolutionServerAddress string, logger *Utilities.Logger, websocketServer *WebsocketServer.Server) *Client {
	return &Client{
		name:                         name,
		logger:                       logger,
		topicResolutionServerAddress: topicResolutionServerAddress,
		randomizer:                   Utilities.NewRandomizer(Utilities.GetSystemTime()),

		websocketServer: websocketServer,

		handleServerMessagesConcurrently: false,
	}
}

// sets the application that the client will use to handle messages
func (client *Client) SetApplication(application Application.Application) {
	client.stateMutex.Lock()
	defer client.stateMutex.Unlock()
	if client.isStarted {
		client.logger.Log("Cannot set application while client is started")
		return
	}
	client.application = application
}

func (client *Client) SetTopicResolutionServerAddress(address string) {
	client.topicResolutionServerAddress = address
}

func (client *Client) GetApplication() Application.Application {
	return client.application
}

func (client *Client) GetWebsocketServer() *WebsocketServer.Server {
	return client.websocketServer
}

func (client *Client) GetTopicResolutionServerAddress() string {
	return client.topicResolutionServerAddress
}

func (client *Client) SetHandleMessagesConcurrently(handleMessagesConcurrently bool) {
	client.handleServerMessagesConcurrentlyMutex.Lock()
	defer client.handleServerMessagesConcurrentlyMutex.Unlock()
	client.handleServerMessagesConcurrently = handleMessagesConcurrently
}

func (client *Client) GetHandleMessagesConcurrently() bool {
	client.handleServerMessagesConcurrentlyMutex.Lock()
	defer client.handleServerMessagesConcurrentlyMutex.Unlock()
	return client.handleServerMessagesConcurrently
}

func (client *Client) Start() error {
	client.stateMutex.Lock()
	client.mapOperationMutex.Lock()
	if client.application == nil {
		return Error.New("Application not set", nil)
	}
	if client.topicResolutionServerAddress == "" {
		return Error.New("Topic resolution server address not set", nil)
	}
	if client.isStarted {
		return Error.New("Client already connected", nil)
	}
	client.topicResolutions = make(map[string]*serverConnection)
	client.messagesWaitingForResponse = make(map[string]chan *Message.Message)
	client.serverConnections = make(map[string]*serverConnection)
	client.stopChannel = make(chan bool)
	client.isStarted = true
	client.mapOperationMutex.Unlock()

	if client.websocketServer != nil {
		err := client.websocketServer.Start()
		if err != nil {
			client.isStarted = false
			close(client.stopChannel)
			return Error.New("Error starting websocket server", err)
		}
	}

	topics := make([]string, 0)
	for topic := range client.application.GetSyncMessageHandlers() {
		topics = append(topics, topic)
	}
	for topic := range client.application.GetAsyncMessageHandlers() {
		topics = append(topics, topic)
	}
	for _, topic := range topics {
		serverConnection, err := client.getServerConnectionForTopic(topic)
		if err != nil {
			panic(err)
		}
		err = client.subscribeTopic(serverConnection, topic)
		if err != nil {
			panic(err)
		}
	}
	client.stateMutex.Unlock()
	err := client.application.OnStart()
	if err != nil {
		client.Stop()
		return Error.New("Error in OnStart", err)
	}
	return nil
}

func (client *Client) Stop() error {
	client.stateMutex.Lock()
	if !client.isStarted {
		return Error.New("Client not connected", nil)
	}
	err := client.application.OnStop()
	if err != nil {
		return Error.New("Error in OnStop", err)
	}
	if client.websocketServer != nil {
		err := client.websocketServer.Stop()
		if err != nil {
			return Error.New("Error stopping websocket server", err)
		}
	}
	client.mapOperationMutex.Lock()
	for _, connection := range client.serverConnections {
		connection.close()
	}
	client.serverConnections = make(map[string]*serverConnection)
	client.topicResolutions = make(map[string]*serverConnection)
	client.messagesWaitingForResponse = make(map[string]chan *Message.Message)
	client.isStarted = false
	close(client.stopChannel)
	client.mapOperationMutex.Unlock()

	client.stateMutex.Unlock()
	return nil
}

func (server *Client) GetName() string {
	return server.name
}
