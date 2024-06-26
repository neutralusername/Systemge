package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Config struct {
	Name       string // *required*
	LoggerPath string // *required*

	ResolverAddress        string // *required*
	ResolverNameIndication string // *required*
	ResolverTLSCert        string // *required*

	HandleMessagesSequentially bool // default: false

	HTTPPort     string // *optional*
	HTTPCertPath string // *optional*
	HTTPKeyPath  string // *optional*

	WebsocketPattern  string // *optional*
	WebsocketPort     string // *optional*
	WebsocketCertPath string // *optional*
	WebsocketKeyPath  string // *optional*
}

type Node struct {
	config *Config

	logger     *Utilities.Logger
	randomizer *Utilities.Randomizer

	stopChannel chan bool
	isStarted   bool

	handleMessagesSequentiallyMutex sync.Mutex
	websocketMutex                  sync.Mutex
	httpMutex                       sync.Mutex
	clientMutex                     sync.Mutex
	stateMutex                      sync.Mutex

	application Application
	//client
	messagesWaitingForResponse map[string]chan *Message.Message // syncKey -> responseChannel
	activeBrokerConnections    map[string]*brokerConnection     // brokerAddress -> serverConnection
	topicResolutions           map[string]*brokerConnection     // topic -> serverConnection

	websocketApplication WebsocketComponent
	//websocket
	websocketHandshakeHTTPServer *http.Server
	websocketConnChannel         chan *websocket.Conn
	websocketClients             map[string]*WebsocketClient            // websocketId -> websocketClient
	WebsocketGroups              map[string]map[string]*WebsocketClient // groupId -> map[websocketId]websocketClient
	websocketClientGroups        map[string]map[string]bool             // websocketId -> map[groupId]bool

	httpApplication HTTPComponent
	//http
	httpServer *http.Server
}

func New(clientConfig *Config, application Application, httpApplication HTTPComponent, websocketApplication WebsocketComponent) *Node {
	return &Node{
		config: clientConfig,

		logger:     Utilities.NewLogger(clientConfig.LoggerPath),
		randomizer: Utilities.NewRandomizer(Utilities.GetSystemTime()),

		application:          application,
		httpApplication:      httpApplication,
		websocketApplication: websocketApplication,

		messagesWaitingForResponse: make(map[string]chan *Message.Message),

		topicResolutions:        make(map[string]*brokerConnection),
		activeBrokerConnections: make(map[string]*brokerConnection),

		WebsocketGroups:       make(map[string]map[string]*WebsocketClient),
		websocketClients:      make(map[string]*WebsocketClient),
		websocketClientGroups: make(map[string]map[string]bool),
	}
}

func (client *Node) Start() error {
	err := func() error {
		client.stateMutex.Lock()
		defer client.stateMutex.Unlock()
		if client.isStarted {
			return Error.New("Client already connected", nil)
		}
		if client.application == nil {
			return Error.New("Application not set", nil)
		}
		if client.websocketApplication != nil {
			err := client.startWebsocketServer()
			if err != nil {
				return Error.New("Error starting websocket server", err)
			}
		}
		if client.httpApplication != nil {
			err := client.startApplicationHTTPServer()
			if err != nil {
				return Error.New("Error starting http server", err)
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
				return Error.New("Error getting server connection for topic", err)
			}
			err = client.subscribeTopic(serverConnection, topic)
			if err != nil {
				client.removeAllBrokerConnections()
				close(client.stopChannel)
				return Error.New("Error subscribing to topic", err)
			}
		}
		client.isStarted = true
		return nil
	}()
	if err != nil {
		return Error.New("Error starting client", err)
	}
	err = client.application.OnStart(client)
	if err != nil {
		client.Stop()
		return Error.New("Error in OnStart", err)
	}
	return nil
}

func (client *Node) Stop() error {
	client.stateMutex.Lock()
	defer client.stateMutex.Unlock()
	if !client.isStarted {
		return Error.New("Client not started", nil)
	}
	err := client.application.OnStop(client)
	if err != nil {
		return Error.New("Error in OnStop", err)
	}
	if client.websocketApplication != nil {
		err := client.stopWebsocketServer()
		if err != nil {
			return Error.New("Error stopping websocket server", err)
		}
	}
	if client.httpApplication != nil {
		err := client.stopApplicationHTTPServer()
		if err != nil {
			return Error.New("Error stopping http server", err)
		}
	}
	client.isStarted = false
	client.removeAllBrokerConnections()
	close(client.stopChannel)
	return nil
}

func (client *Node) IsStarted() bool {
	client.stateMutex.Lock()
	defer client.stateMutex.Unlock()
	return client.isStarted
}
