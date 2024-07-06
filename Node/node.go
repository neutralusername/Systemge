package Node

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Node struct {
	config Config.Node

	logger     *Utilities.Logger
	randomizer *Utilities.Randomizer

	application        Application
	httpComponent      HTTPComponent
	websocketComponent WebsocketComponent

	stopChannel chan bool
	isStarted   bool

	handleMessagesSequentiallyMutex sync.Mutex
	websocketMutex                  sync.Mutex
	httpMutex                       sync.Mutex
	mutex                           sync.Mutex
	stateMutex                      sync.Mutex

	messagesWaitingForResponse map[string]chan *Message.Message // syncKey -> responseChannel
	brokerConnections          map[string]*brokerConnection     // brokerAddress -> brokerConnection
	topicBrokerConnections     map[string]*brokerConnection     // topic -> brokerConnection

	//websocket
	websocketHandshakeHTTPServer *http.Server
	websocketConnChannel         chan *websocket.Conn
	websocketClients             map[string]*WebsocketClient            // websocketId -> websocketClient
	WebsocketGroups              map[string]map[string]*WebsocketClient // groupId -> map[websocketId]websocketClient
	websocketClientGroups        map[string]map[string]bool             // websocketId -> map[groupId]bool

	//http
	httpServer *http.Server
}

func New(config Config.Node, application Application) *Node {
	node := &Node{
		config: config,

		logger:     Utilities.NewLogger(config.LoggerPath),
		randomizer: Utilities.NewRandomizer(Utilities.GetSystemTime()),

		application: application,

		messagesWaitingForResponse: make(map[string]chan *Message.Message),
		brokerConnections:          make(map[string]*brokerConnection),
		topicBrokerConnections:     make(map[string]*brokerConnection),

		WebsocketGroups:       make(map[string]map[string]*WebsocketClient),
		websocketClients:      make(map[string]*WebsocketClient),
		websocketClientGroups: make(map[string]map[string]bool),
	}
	if httpComponent, ok := application.(HTTPComponent); ok {
		node.httpComponent = httpComponent
	}
	if websocketComponent, ok := application.(WebsocketComponent); ok {
		node.websocketComponent = websocketComponent
	}
	return node
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
			go node.subscribeLoop(topic)
		}
		node.isStarted = true
		return nil
	}()
	if err != nil {
		return Error.New("Error starting node \""+node.GetName()+"\"", err)
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
	close(node.stopChannel)
	node.removeAllBrokerConnections()
	return nil
}

func (node *Node) IsStarted() bool {
	node.stateMutex.Lock()
	defer node.stateMutex.Unlock()
	return node.isStarted
}
