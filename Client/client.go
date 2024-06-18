package Client

import (
	"Systemge/Application"
	"Systemge/Message"
	"Systemge/Randomizer"
	"Systemge/Utilities"
	"Systemge/WebsocketServer"
	"sync"
)

type Client struct {
	name            string
	logger          *Utilities.Logger
	resolverAddress string
	randomizer      *Randomizer.Randomizer

	websocketServer *WebsocketServer.Server
	application     Application.Application

	messagesWaitingForResponse map[string]chan *Message.Message // syncKey -> responseChannel

	activeBrokerConnections map[string]*brokerConnection // brokerAddress -> serverConnection
	topicResolutions        map[string]*brokerConnection // topic -> serverConnection
	mapOperationMutex       sync.Mutex

	// handleMessagesConcurrently is a flag that determines whether the client will handle messages by brokers concurrently
	// If this is set to true, the client will handle messages concurrently, otherwise it will handle messages sequentially
	// concurrently handling messages can be useful to improve performance but it will often require additional work for the application to be able to handle concurrency
	handleMessagesConcurrently      bool
	handleMessagesConcurrentlyMutex sync.Mutex

	stopChannel chan bool
	isStarted   bool

	stateMutex sync.Mutex
}

func New(name, topicResolutionServerAddress string, logger *Utilities.Logger, websocketServer *WebsocketServer.Server) *Client {
	return &Client{
		name:            name,
		logger:          logger,
		resolverAddress: topicResolutionServerAddress,
		randomizer:      Randomizer.New(Randomizer.GetSystemTime()),

		messagesWaitingForResponse: make(map[string]chan *Message.Message),
		topicResolutions:           make(map[string]*brokerConnection),
		activeBrokerConnections:    make(map[string]*brokerConnection),

		websocketServer: websocketServer,

		handleMessagesConcurrently: true,
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
	client.resolverAddress = address
}

func (client *Client) GetApplication() Application.Application {
	return client.application
}

func (client *Client) GetWebsocketServer() *WebsocketServer.Server {
	return client.websocketServer
}

func (client *Client) GetTopicResolutionServerAddress() string {
	return client.resolverAddress
}

func (client *Client) GetLogger() *Utilities.Logger {
	return client.logger
}

func (client *Client) SetLogger(logger *Utilities.Logger) {
	client.logger = logger
}

func (client *Client) SetHandleMessagesConcurrently(handleMessagesConcurrently bool) {
	client.handleMessagesConcurrentlyMutex.Lock()
	defer client.handleMessagesConcurrentlyMutex.Unlock()
	client.handleMessagesConcurrently = handleMessagesConcurrently
}

func (client *Client) GetHandleMessagesConcurrently() bool {
	client.handleMessagesConcurrentlyMutex.Lock()
	defer client.handleMessagesConcurrentlyMutex.Unlock()
	return client.handleMessagesConcurrently
}

func (client *Client) Start() error {
	err := func() error {
		client.stateMutex.Lock()
		defer client.stateMutex.Unlock()
		if client.application == nil {
			return Utilities.NewError("Application not set", nil)
		}
		if client.resolverAddress == "" {
			return Utilities.NewError("Topic resolution server address not set", nil)
		}
		if client.isStarted {
			return Utilities.NewError("Client already connected", nil)
		}
		if client.websocketServer != nil {
			err := client.websocketServer.Start()
			if err != nil {
				return Utilities.NewError("Error starting websocket server", err)
			}
		}
		client.stopChannel = make(chan bool)
		topics := make([]string, 0)
		for topic := range client.application.GetSyncMessageHandlers() {
			topics = append(topics, topic)
		}
		for topic := range client.application.GetAsyncMessageHandlers() {
			topics = append(topics, topic)
		}
		for _, topic := range topics {
			serverConnection, err := client.getBrokerConnectionForTopic(topic)
			if err != nil {
				close(client.stopChannel)
				return Utilities.NewError("Error getting server connection for topic", err)
			}
			err = client.subscribeTopic(serverConnection, topic)
			if err != nil {
				close(client.stopChannel)
				return Utilities.NewError("Error subscribing to topic", err)
			}
		}
		client.isStarted = true
		return nil
	}()
	if err != nil {
		return Utilities.NewError("Error starting client", err)
	}

	err = client.application.OnStart()
	if err != nil {
		client.Stop()
		return Utilities.NewError("Error in OnStart", err)
	}
	return nil
}

func (client *Client) Stop() error {
	client.stateMutex.Lock()
	defer client.stateMutex.Unlock()
	if !client.isStarted {
		return Utilities.NewError("Client not started", nil)
	}
	err := client.application.OnStop()
	if err != nil {
		return Utilities.NewError("Error in OnStop", err)
	}
	if client.websocketServer != nil {
		err := client.websocketServer.Stop()
		if err != nil {
			return Utilities.NewError("Error stopping websocket server", err)
		}
	}
	client.mapOperationMutex.Lock()
	for address, brokerConnection := range client.activeBrokerConnections {
		brokerConnection.close()
		delete(client.activeBrokerConnections, address)
		for topic := range brokerConnection.topics {
			delete(client.topicResolutions, topic)
		}
	}
	close(client.stopChannel)
	client.isStarted = false
	client.mapOperationMutex.Unlock()
	return nil
}

func (server *Client) GetName() string {
	return server.name
}
