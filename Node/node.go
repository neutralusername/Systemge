package Node

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Helpers"
	"Systemge/Http"
	"Systemge/Message"
	"Systemge/Tools"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Node struct {
	ongoingSubscribeLoops int
	name                  string
	randomizer            *Tools.Randomizer
	errorLogger           *Tools.Logger
	warningLogger         *Tools.Logger
	infoLogger            *Tools.Logger
	debugLogger           *Tools.Logger
	mailer                *Tools.Mailer

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
	websocketStarted      bool
	websocketMutex        sync.Mutex
	websocketHttpServer   *Http.Server
	websocketConnChannel  chan *websocket.Conn
	websocketClients      map[string]*WebsocketClient            // websocketId -> websocketClient
	websocketGroups       map[string]map[string]*WebsocketClient // groupId -> map[websocketId]websocketClient
	websocketClientGroups map[string]map[string]bool             // websocketId -> map[groupId]bool

	//http
	httpStarted bool
	httpMutex   sync.Mutex
	httpServer  *Http.Server
}

func New(config *Config.Node, application Application) *Node {
	node := &Node{
		name:          config.Name,
		errorLogger:   Tools.NewLogger(config.ErrorLogger),
		warningLogger: Tools.NewLogger(config.WarningLogger),
		infoLogger:    Tools.NewLogger(config.InfoLogger),
		debugLogger:   Tools.NewLogger(config.DebugLogger),
		mailer:        Tools.NewMailer(config.Mailer),

		application: application,

		randomizer: Tools.NewRandomizer(config.RandomizerSeed),

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
	if infoLogger := node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Starting", nil).Error())
	}

	node.stopChannel = make(chan bool)
	node.isStarted = true

	if ImplementsSystemgeComponent(node.application) {
		err := node.startSystemgeComponent()
		if err != nil {
			if err := node.stop(false); err != nil {
				node.GetWarningLogger().Log(Error.New("failed to stop node. Error in OnStart", err).Error())
			}
			return Error.New("failed starting systemge component", err)
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Started systemge component", nil).Error())
		}
	}
	if ImplementsWebsocketComponent(node.application) {
		err := node.startWebsocketComponent()
		if err != nil {
			if err := node.stop(false); err != nil {
				node.GetWarningLogger().Log(Error.New("failed to stop node. Error in OnStart", err).Error())
			}
			return Error.New("failed starting websocket server", err)
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Started websocket component", nil).Error())
		}
	}
	if ImplementsHTTPComponent(node.application) {
		err := node.startHTTPComponent()
		if err != nil {
			if err := node.stop(false); err != nil {
				node.GetWarningLogger().Log(Error.New("failed to stop node. Error in OnStart", err).Error())
			}
			return Error.New("failed starting http server", err)
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Started http component", nil).Error())
		}
	}
	if ImplementsOnStartComponent(node.application) {
		err := node.GetOnStartComponent().OnStart(node)
		if err != nil {
			if err := node.stop(false); err != nil {
				node.GetWarningLogger().Log(Error.New("failed to stop node. Error in OnStart", err).Error())
			}
			return Error.New("failed in OnStart", err)
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("executed OnStart", nil).Error())
		}
	}
	if infoLogger := node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Started", nil).Error())
	}
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
	if infoLogger := node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Stopping", nil).Error())
	}
	if ImplementsOnStopComponent(node.application) {
		err := node.GetOnStopComponent().OnStop(node)
		if err != nil {
			return Error.New("failed to stop node. Error in OnStop", err)
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("executed OnStop", nil).Error())
		}
	}
	if node.websocketStarted {
		err := node.stopWebsocketComponent()
		if err != nil {
			return Error.New("failed to stop node. Error stopping websocket server", err)
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopped websocket component", nil).Error())
		}
	}
	if node.httpStarted {
		err := node.stopHTTPComponent()
		if err != nil {
			return Error.New("failed to stop node. Error stopping http server", err)
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopped http component", nil).Error())
		}
	}
	if node.systemgeStarted {
		err := node.stopSystemgeComponent()
		if err != nil {
			return Error.New("failed to stop node. Error stopping systemge component", err)
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopped systemge component", nil).Error())
		}
	}
	node.isStarted = false
	close(node.stopChannel)
	for node.ongoingSubscribeLoops > 0 {
		node.GetWarningLogger().Log(Error.New("Waiting for "+Helpers.IntToString(node.ongoingSubscribeLoops)+" subscribe loops to finish", nil).Error())
		time.Sleep(time.Duration(node.GetSystemgeComponent().GetSystemgeComponentConfig().BrokerSubscribeDelayMs) * time.Millisecond)
	}
	if infoLogger := node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Stopped", nil).Error())
	}
	return nil
}

func (node *Node) IsStarted() bool {
	return node.isStarted
}

func (node *Node) GetName() string {
	return node.name
}

func (node *Node) GetErrorLogger() *Tools.Logger {
	return node.errorLogger
}

func (node *Node) SetErrorLogger(logger *Tools.Logger) {
	node.errorLogger = logger
}

func (node *Node) GetWarningLogger() *Tools.Logger {
	return node.warningLogger
}

func (node *Node) SetWarningLogger(logger *Tools.Logger) {
	node.warningLogger = logger
}

func (node *Node) GetInfoLogger() *Tools.Logger {
	return node.infoLogger
}

func (node *Node) SetInfoLogger(logger *Tools.Logger) {
	node.infoLogger = logger
}

func (node *Node) GetDebugLogger() *Tools.Logger {
	return node.debugLogger
}

func (node *Node) SetDebugLogger(logger *Tools.Logger) {
	node.debugLogger = logger
}

func (node *Node) GetMailer() *Tools.Mailer {
	return node.mailer
}

func (node *Node) SetMailer(mailer *Tools.Mailer) {
	node.mailer = mailer
}

func (node *Node) GetRandomizer() *Tools.Randomizer {
	return node.randomizer
}

func (node *Node) SetRandomizer(randomizer *Tools.Randomizer) {
	node.randomizer = randomizer
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
