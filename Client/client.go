package Client

import (
	"Systemge/Message"
	"Systemge/Randomizer"
	"Systemge/Utilities"
	"Systemge/WebsocketClient"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Config struct {
	Name       string
	LoggerPath string

	HandleMessagesConcurrently bool

	ResolverAddress        string
	ResolverNameIndication string
	ResolverTLSCert        string

	HTTPPort     string
	HTTPCertPath string
	HTTPKeyPath  string

	WebsocketPattern  string
	WebsocketPort     string
	WebsocketCertPath string
	WebsocketKeyPath  string
}

type Client struct {
	config *Config

	logger     *Utilities.Logger
	randomizer *Randomizer.Randomizer

	stopChannel chan bool
	isStarted   bool

	handleMessagesConcurrentlyMutex sync.Mutex
	websocketMutex                  sync.Mutex
	httpMutex                       sync.Mutex
	clientMutex                     sync.Mutex
	stateMutex                      sync.Mutex

	application Application
	//client
	messagesWaitingForResponse map[string]chan *Message.Message // syncKey -> responseChannel
	activeBrokerConnections    map[string]*brokerConnection     // brokerAddress -> serverConnection
	topicResolutions           map[string]*brokerConnection     // topic -> serverConnection

	websocketApplication WebsocketApplication
	//websocket
	websocketHandshakeHTTPServer *http.Server
	websocketConnChannel         chan *websocket.Conn
	websocketClients             map[string]*WebsocketClient.Client            // websocketId -> websocketClient
	groups                       map[string]map[string]*WebsocketClient.Client // groupId -> map[websocketId]websocketClient
	websocketClientGroups        map[string]map[string]bool                    // websocketId -> map[groupId]bool

	httpApplication HTTPApplication
	//http
	httpServer *http.Server
}

func New(clientConfig *Config, application Application, httpApplication HTTPApplication, websocketApplication WebsocketApplication) *Client {
	return &Client{
		config: clientConfig,

		logger:     Utilities.NewLogger(clientConfig.LoggerPath),
		randomizer: Randomizer.New(Randomizer.GetSystemTime()),

		application:          application,
		httpApplication:      httpApplication,
		websocketApplication: websocketApplication,

		messagesWaitingForResponse: make(map[string]chan *Message.Message),

		topicResolutions:        make(map[string]*brokerConnection),
		activeBrokerConnections: make(map[string]*brokerConnection),

		groups:                make(map[string]map[string]*WebsocketClient.Client),
		websocketClients:      make(map[string]*WebsocketClient.Client),
		websocketClientGroups: make(map[string]map[string]bool),
	}
}

func (client *Client) Start() error {
	err := func() error {
		client.stateMutex.Lock()
		defer client.stateMutex.Unlock()
		if client.isStarted {
			return Utilities.NewError("Client already connected", nil)
		}
		if client.application == nil {
			return Utilities.NewError("Application not set", nil)
		}
		if client.websocketApplication != nil {
			err := client.StartWebsocketServer()
			if err != nil {
				return Utilities.NewError("Error starting websocket server", err)
			}
		}
		if client.httpApplication != nil {
			err := client.StartApplicationHTTPServer()
			if err != nil {
				return Utilities.NewError("Error starting http server", err)
			}
		}
		client.stopChannel = make(chan bool)
		topicsToSubscribeTo := []string{}
		for topic := range client.application.GetAsyncMessageHandlers() {
			topicsToSubscribeTo = append(topicsToSubscribeTo, topic)
		}
		for topic := range client.application.GetSyncMessageHandlers() {
			topicsToSubscribeTo = append(topicsToSubscribeTo, topic)
		}
		for _, topic := range topicsToSubscribeTo {
			serverConnection, err := client.getBrokerConnectionForTopic(topic)
			if err != nil {
				client.removeAllBrokerConnections()
				close(client.stopChannel)
				return Utilities.NewError("Error getting server connection for topic", err)
			}
			err = client.subscribeTopic(serverConnection, topic)
			if err != nil {
				client.removeAllBrokerConnections()
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
	err = client.application.OnStart(client)
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
	err := client.application.OnStop(client)
	if err != nil {
		return Utilities.NewError("Error in OnStop", err)
	}
	if client.websocketApplication != nil {
		err := client.StopWebsocketServer()
		if err != nil {
			return Utilities.NewError("Error stopping websocket server", err)
		}
	}
	if client.httpApplication != nil {
		err := client.StopApplicationHTTPServer()
		if err != nil {
			return Utilities.NewError("Error stopping http server", err)
		}
	}
	client.isStarted = false
	client.removeAllBrokerConnections()
	close(client.stopChannel)
	return nil
}

func (client *Client) IsStarted() bool {
	client.stateMutex.Lock()
	defer client.stateMutex.Unlock()
	return client.isStarted
}
