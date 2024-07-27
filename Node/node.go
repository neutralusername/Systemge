package Node

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tools"
	"sync"
)

type Node struct {
	config     *Config.Node
	randomizer *Tools.Randomizer

	stopChannel chan bool
	isStarted   bool
	mutex       sync.Mutex

	application Application

	//systemge
	systemge *systemgeComponent

	//websocket
	websocket *websocketComponent

	//http
	http *httpComponent
}

func New(config *Config.Node, application Application) *Node {
	node := &Node{
		config: config,

		application: application,

		randomizer: Tools.NewRandomizer(config.RandomizerSeed),
	}
	return node
}

func (node *Node) StartBlocking() error {
	if node.isStarted {
		return Error.New("node already started", nil)
	}
	err := node.Start()
	<-node.stopChannel
	return err
}

func (node *Node) Start() error {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	if node.IsStarted() {
		return Error.New("node already started", nil)
	}
	if infoLogger := node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Starting", nil).Error())
	}
	node.stopChannel = make(chan bool)
	node.isStarted = true
	if ImplementsSystemgeComponent(node.application) {
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Starting systemge component", nil).Error())
		}
		err := node.startSystemgeComponent()
		if err != nil {
			if err := node.stop(false); err != nil {
				node.GetWarningLogger().Log(Error.New("failed to stop node", err).Error())
			}
			return Error.New("failed starting systemge component", err)
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Started systemge component", nil).Error())
		}
	}
	if ImplementsWebsocketComponent(node.application) {
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Starting websocket component", nil).Error())
		}
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
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Starting http component", nil).Error())
		}
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
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Executing OnStart", nil).Error())
		}
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
		node.mutex.Lock()
		defer node.mutex.Unlock()
	}
	if !node.IsStarted() {
		return Error.New("node not started", nil)
	}
	if infoLogger := node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Stopping", nil).Error())
	}
	if ImplementsOnStopComponent(node.application) {
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Executing OnStop", nil).Error())
		}
		err := node.GetOnStopComponent().OnStop(node)
		if err != nil {
			return Error.New("failed to stop node. Error in OnStop", err)
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("executed OnStop", nil).Error())
		}
	}
	if node.websocket != nil {
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopping websocket component", nil).Error())
		}
		err := node.stopWebsocketComponent()
		if err != nil {
			return Error.New("failed to stop node. Error stopping websocket server", err)
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopped websocket component", nil).Error())
		}
	}
	if node.http != nil {
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopping http component", nil).Error())
		}
		err := node.stopHTTPComponent()
		if err != nil {
			return Error.New("failed to stop node. Error stopping http server", err)
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopped http component", nil).Error())
		}
	}
	if node.systemge != nil {
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopping systemge component", nil).Error())
		}
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
	if infoLogger := node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Stopped", nil).Error())
	}
	return nil
}

func (node *Node) IsStarted() bool {
	return node.isStarted
}

func (node *Node) GetName() string {
	return node.config.Name
}

func (node *Node) GetErrorLogger() *Tools.Logger {
	return node.config.ErrorLogger
}

func (node *Node) GetWarningLogger() *Tools.Logger {
	return node.config.WarningLogger
}

func (node *Node) GetInfoLogger() *Tools.Logger {
	return node.config.InfoLogger
}

func (node *Node) GetDebugLogger() *Tools.Logger {
	return node.config.DebugLogger
}

func (node *Node) GetMailer() *Tools.Mailer {
	return node.config.Mailer
}

func (node *Node) GetRandomizer() *Tools.Randomizer {
	return node.randomizer
}

func (node *Node) SetRandomizer(randomizer *Tools.Randomizer) {
	node.randomizer = randomizer
}

func (node *Node) GetCommandHandlerComponent() CommandComponent {
	if ImplementsCommandHandlerComponent(node.application) {
		return node.application.(CommandComponent)
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
