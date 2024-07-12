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

	randomizer *Utilities.Randomizer

	application        Application
	httpComponent      HTTPComponent
	websocketComponent WebsocketComponent

	stopChannel      chan bool
	isStarted        bool
	websocketStarted bool
	httpStarted      bool

	handleMessagesSequentiallyMutex sync.Mutex
	websocketMutex                  sync.Mutex
	httpMutex                       sync.Mutex
	mutex                           sync.Mutex
	stateChangeMutex                sync.Mutex

	messagesWaitingForResponse map[string]chan *Message.Message // syncKey -> responseChannel
	brokerConnections          map[string]*brokerConnection     // brokerAddress -> brokerConnection
	topicResolutions           map[string]*brokerConnection     // topic -> brokerConnection

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

		randomizer: Utilities.NewRandomizer(Utilities.GetSystemTime()),

		application: application,

		messagesWaitingForResponse: make(map[string]chan *Message.Message),
		brokerConnections:          make(map[string]*brokerConnection),
		topicResolutions:           make(map[string]*brokerConnection),

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
	node.stateChangeMutex.Lock()
	defer node.stateChangeMutex.Unlock()
	if node.isStarted {
		return Error.New("node already started", nil)
	}
	if node.application == nil {
		return Error.New("application not set", nil)
	}

	node.stopChannel = make(chan bool)
	node.isStarted = true

	for topic := range node.application.GetAsyncMessageHandlers() {
		node.subscribeLoop(topic)
	}
	for topic := range node.application.GetSyncMessageHandlers() {
		node.subscribeLoop(topic)
	}
	if node.websocketComponent != nil {
		err := node.startWebsocketComponent()
		if err != nil {
			node.stop(false)
			return Error.New("failed starting websocket server", err)
		} else {
			node.config.Logger.Info(Error.New("Started websocket component on node \""+node.GetName()+"\"", nil).Error())
		}
	}
	if node.httpComponent != nil {
		err := node.startHTTPComponent()
		if err != nil {
			node.stop(false)
			return Error.New("failed starting http server", err)
		} else {
			node.config.Logger.Info(Error.New("Started http component on node \""+node.GetName()+"\"", nil).Error())
		}
	}
	err := node.application.OnStart(node)
	if err != nil {
		node.stop(false)
		return Error.New("failed in OnStart", err)
	}
	node.config.Logger.Info(Error.New("Started node \""+node.config.Name+"\"", nil).Error())
	return nil
}

func (node *Node) Stop() error {
	return node.stop(true)
}

func (node *Node) stop(lock bool) error {
	if lock {
		node.stateChangeMutex.Lock()
		defer node.stateChangeMutex.Unlock()
	}
	if !node.isStarted {
		return Error.New("node not started", nil)
	}
	err := node.application.OnStop(node)
	if err != nil {
		return Error.New("failed to stop node. Error in OnStop", err)
	}
	if node.websocketStarted {
		err := node.stopWebsocketComponent()
		if err != nil {
			return Error.New("failed to stop node. Error stopping websocket server", err)
		} else {
			node.config.Logger.Info(Error.New("Stopped websocket component on node \""+node.GetName()+"\"", nil).Error())
		}
	}
	if node.httpStarted {
		err := node.stopHTTPComponent()
		if err != nil {
			return Error.New("failed to stop node. Error stopping http server", err)
		} else {
			node.config.Logger.Info(Error.New("Stopped http component on node \""+node.GetName()+"\"", nil).Error())
		}
	}
	node.isStarted = false
	close(node.stopChannel)
	node.removeAllBrokerConnections()
	node.config.Logger.Info(Error.New("Stopped node \""+node.config.Name+"\"", nil).Error())
	return nil
}

func (node *Node) IsStarted() bool {
	return node.isStarted
}
