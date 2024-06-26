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
	mutex                           sync.Mutex
	stateMutex                      sync.Mutex

	application Application
	//basic
	messagesWaitingForResponse map[string]chan *Message.Message // syncKey -> responseChannel
	activeBrokerConnections    map[string]*brokerConnection     // brokerAddress -> serverConnection
	topicResolutions           map[string]*brokerConnection     // topic -> serverConnection

	websocketComponent WebsocketComponent
	//websocket
	websocketHandshakeHTTPServer *http.Server
	websocketConnChannel         chan *websocket.Conn
	websocketClients             map[string]*WebsocketClient            // websocketId -> websocketClient
	WebsocketGroups              map[string]map[string]*WebsocketClient // groupId -> map[websocketId]websocketClient
	websocketClientGroups        map[string]map[string]bool             // websocketId -> map[groupId]bool

	httpComponent HTTPComponent
	//http
	httpServer *http.Server
}

func New(config *Config, application Application, httpComponent HTTPComponent, websocketComponent WebsocketComponent) *Node {
	return &Node{
		config: config,

		logger:     Utilities.NewLogger(config.LoggerPath),
		randomizer: Utilities.NewRandomizer(Utilities.GetSystemTime()),

		application:        application,
		httpComponent:      httpComponent,
		websocketComponent: websocketComponent,

		messagesWaitingForResponse: make(map[string]chan *Message.Message),

		topicResolutions:        make(map[string]*brokerConnection),
		activeBrokerConnections: make(map[string]*brokerConnection),

		WebsocketGroups:       make(map[string]map[string]*WebsocketClient),
		websocketClients:      make(map[string]*WebsocketClient),
		websocketClientGroups: make(map[string]map[string]bool),
	}
}

func (node *Node) Start() error {
	err := func() error {
		node.stateMutex.Lock()
		defer node.stateMutex.Unlock()
		if node.isStarted {
			return Error.New("Node already started", nil)
		}
		if node.application == nil {
			return Error.New("Application not set", nil)
		}
		if node.websocketComponent != nil {
			err := node.startWebsocketComponent()
			if err != nil {
				return Error.New("Error starting websocket server", err)
			}
		}
		if node.httpComponent != nil {
			err := node.startHTTPComponent()
			if err != nil {
				return Error.New("Error starting http server", err)
			}
		}
		node.stopChannel = make(chan bool)
		topicsToSubscribeTo := []string{}
		for topic := range node.application.GetAsyncMessageHandlers() {
			topicsToSubscribeTo = append(topicsToSubscribeTo, topic)
		}
		for topic := range node.application.GetSyncMessageHandlers() {
			topicsToSubscribeTo = append(topicsToSubscribeTo, topic)
		}
		for _, topic := range topicsToSubscribeTo {
			serverConnection, err := node.getBrokerConnectionForTopic(topic)
			if err != nil {
				node.removeAllBrokerConnections()
				close(node.stopChannel)
				return Error.New("Error getting server connection for topic", err)
			}
			err = node.subscribeTopic(serverConnection, topic)
			if err != nil {
				node.removeAllBrokerConnections()
				close(node.stopChannel)
				return Error.New("Error subscribing to topic", err)
			}
		}
		node.isStarted = true
		return nil
	}()
	if err != nil {
		return Error.New("Error starting node", err)
	}
	err = node.application.OnStart(node)
	if err != nil {
		node.Stop()
		return Error.New("Error in OnStart", err)
	}
	return nil
}

func (node *Node) Stop() error {
	node.stateMutex.Lock()
	defer node.stateMutex.Unlock()
	if !node.isStarted {
		return Error.New("Node not started", nil)
	}
	err := node.application.OnStop(node)
	if err != nil {
		return Error.New("Error in OnStop", err)
	}
	if node.websocketComponent != nil {
		err := node.stopWebsocketComponent()
		if err != nil {
			return Error.New("Error stopping websocket server", err)
		}
	}
	if node.httpComponent != nil {
		err := node.stopHTTPComponent()
		if err != nil {
			return Error.New("Error stopping http server", err)
		}
	}
	node.isStarted = false
	node.removeAllBrokerConnections()
	close(node.stopChannel)
	return nil
}

func (node *Node) IsStarted() bool {
	node.stateMutex.Lock()
	defer node.stateMutex.Unlock()
	return node.isStarted
}
