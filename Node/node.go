package Node

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tools"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Node struct {
	config     Config.Node
	randomizer *Tools.Randomizer
	logger     *Tools.Logger
	mailer     *Tools.Mailer

	stopChannel chan bool
	isStarted   bool
	nodeMutex   sync.Mutex

	application Application

	//systemge
	systemgeStarted                    bool
	systemgeMutex                      sync.Mutex
	systemgeHandleSequentiallyMutex    sync.Mutex
	systemgeMessagesWaitingForResponse map[string]chan *Message.Message // syncKey -> responseChannel
	systemgeBrokerConnections          map[string]*brokerConnection     // brokerAddress -> brokerConnection
	systemgeTopicResolutions           map[string]*brokerConnection     // topic -> brokerConnection

	//websocket
	websocketStarted             bool
	websocketMutex               sync.Mutex
	websocketHandshakeHTTPServer *http.Server
	websocketConnChannel         chan *websocket.Conn
	websocketClients             map[string]*WebsocketClient            // websocketId -> websocketClient
	websocketGroups              map[string]map[string]*WebsocketClient // groupId -> map[websocketId]websocketClient
	websocketClientGroups        map[string]map[string]bool             // websocketId -> map[groupId]bool

	//http
	httpStarted bool
	httpMutex   sync.Mutex
	httpServer  *http.Server
}

func New(config Config.Node, application Application) *Node {
	node := &Node{
		config: config,
		logger: Tools.NewLogger(config.Logger),
		mailer: Tools.NewMailer(config.Mailer),

		application: application,

		randomizer: Tools.NewRandomizer(Tools.GetSystemTime()),

		systemgeMessagesWaitingForResponse: make(map[string]chan *Message.Message),
		systemgeBrokerConnections:          make(map[string]*brokerConnection),
		systemgeTopicResolutions:           make(map[string]*brokerConnection),

		websocketGroups:       make(map[string]map[string]*WebsocketClient),
		websocketConnChannel:  make(chan *websocket.Conn),
		websocketClients:      make(map[string]*WebsocketClient),
		websocketClientGroups: make(map[string]map[string]bool),
	}
	return node
}

func (node *Node) Start() error {
	node.nodeMutex.Lock()
	defer node.nodeMutex.Unlock()
	if node.IsStarted() {
		return Error.New("node already started", nil)
	}
	if node.application == nil {
		return Error.New("application not set", nil)
	}

	node.stopChannel = make(chan bool)
	node.isStarted = true

	if ImplementsSystemgeComponent(node.application) {
		err := node.startSystemgeComponent()
		if err != nil {
			node.stop(false)
			return Error.New("failed starting systemge component", err)
		}
		node.GetLogger().Info(Error.New("Started systemge component on node \""+node.GetName()+"\"", nil).Error())
	}
	if ImplementsWebsocketComponent(node.application) {
		err := node.startWebsocketComponent()
		if err != nil {
			node.stop(false)
			return Error.New("failed starting websocket server", err)
		}
		node.GetLogger().Info(Error.New("Started websocket component on node \""+node.GetName()+"\"", nil).Error())
	}
	if ImplementsHTTPComponent(node.application) {
		err := node.startHTTPComponent()
		if err != nil {
			node.stop(false)
			return Error.New("failed starting http server", err)
		}
		node.GetLogger().Info(Error.New("Started http component on node \""+node.GetName()+"\"", nil).Error())
	}
	if ImplementsOnStartComponent(node.application) {
		err := node.GetOnStartComponent().OnStart(node)
		if err != nil {
			node.stop(false)
			return Error.New("failed in OnStart", err)
		}
		node.GetLogger().Info(Error.New("executed OnStart on node \""+node.GetName()+"\"", nil).Error())
	}
	node.GetLogger().Info(Error.New("Started node \""+node.GetName()+"\"", nil).Error())
	return nil
}

func (node *Node) Stop() error {
	return node.stop(true)
}

func (node *Node) stop(lock bool) error {
	if lock {
		node.nodeMutex.Lock()
		defer node.nodeMutex.Unlock()
	}
	if !node.IsStarted() {
		return Error.New("node not started", nil)
	}
	if ImplementsOnStopComponent(node.application) {
		err := node.GetOnStopComponent().OnStop(node)
		if err != nil {
			return Error.New("failed to stop node. Error in OnStop", err)
		}
		node.GetLogger().Info(Error.New("executed OnStop on node \""+node.GetName()+"\"", nil).Error())
	}
	if node.websocketStarted {
		err := node.stopWebsocketComponent()
		if err != nil {
			return Error.New("failed to stop node. Error stopping websocket server", err)
		}
		node.GetLogger().Info(Error.New("Stopped websocket component on node \""+node.GetName()+"\"", nil).Error())
	}
	if node.httpStarted {
		err := node.stopHTTPComponent()
		if err != nil {
			return Error.New("failed to stop node. Error stopping http server", err)
		}
		node.GetLogger().Info(Error.New("Stopped http component on node \""+node.GetName()+"\"", nil).Error())
	}
	if node.systemgeStarted {
		err := node.stopSystemgeComponent()
		if err != nil {
			return Error.New("failed to stop node. Error stopping systemge component", err)
		}
		node.GetLogger().Info(Error.New("Stopped systemge component on node \""+node.GetName()+"\"", nil).Error())
	}
	node.isStarted = false
	close(node.stopChannel)
	node.GetLogger().Info(Error.New("Stopped node \""+node.GetName()+"\"", nil).Error())
	return nil
}

func (node *Node) IsStarted() bool {
	return node.isStarted
}

func (node *Node) GetName() string {
	return node.config.Name
}

func (node *Node) GetLogger() *Tools.Logger {
	return node.logger
}

func (node *Node) SetLogger(logger *Tools.Logger) {
	node.logger = logger
}

func (node *Node) GetMailer() *Tools.Mailer {
	return node.mailer
}

func (node *Node) SetMailer(mailer *Tools.Mailer) {
	node.mailer = mailer
}

func (node *Node) GetSystemgeComponent() SystemgeComponent {
	if ImplementsSystemgeComponent(node.application) {
		return node.application.(SystemgeComponent)
	}
	return nil
}

func (node *Node) GetWebsocketComponent() WebsocketComponent {
	if ImplementsWebsocketComponent(node.application) {
		return node.application.(WebsocketComponent)
	}
	return nil
}

func (node *Node) GetHTTPComponent() HTTPComponent {
	if ImplementsHTTPComponent(node.application) {
		return node.application.(HTTPComponent)
	}
	return nil
}

func (node *Node) GetCommandHandlerComponent() CommandHandlerComponent {
	if ImplementsCommandHandlerComponent(node.application) {
		return node.application.(CommandHandlerComponent)
	}
	return nil
}

func (node *Node) GetOnStartComponent() OnStartComponent {
	if ImplementsOnStartComponent(node.application) {
		return node.application.(OnStartComponent)
	}
	return nil
}

func (node *Node) GetOnStopComponent() OnStopComponent {
	if ImplementsOnStopComponent(node.application) {
		return node.application.(OnStopComponent)
	}
	return nil
}
