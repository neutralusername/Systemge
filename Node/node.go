package Node

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tools"
)

type Node struct {
	config *Config.Node

	stopChannel chan bool
	status      int
	startMutex  sync.Mutex
	stopMutex   sync.Mutex

	randomizer    *Tools.Randomizer
	mailer        *Tools.Mailer
	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger

	application    Application
	systemgeClient *systemgeClientComponent
	systemgeServer *systemgeServerComponent
}

const (
	STATUS_STOPPED  = 0
	STATUS_STARTING = 1
	STATUS_STARTED  = 2
	STATUS_STOPPING = 3
)

func New(config *Config.Node, application Application) *Node {
	if config == nil {
		panic("NewNodeConfig is required")
	}
	if application == nil {
		panic("Application is required")
	}
	if config.SystemgeServerConfig == nil && config.SystemgeClientConfig == nil {
		panic("At least one of the following configurations is required: HttpConfig, WebsocketConfig, SystemgeConfig")
	}
	implementsSystemgeComponent := ImplementsSystemgeServerComponent(application)
	if !implementsSystemgeComponent && config.SystemgeServerConfig != nil {
		panic("SystemgeConfig provided but application does not implement SystemgeComponent")
	}
	if implementsSystemgeComponent && config.SystemgeServerConfig == nil {
		panic("Application implements SystemgeComponent but SystemgeConfig is missing")
	}
	if implementsSystemgeComponent && config.SystemgeServerConfig.ServerConfig == nil {
		panic("Application implements SystemgeComponent but SystemgeConfig.ServerConfig is missing")
	}
	node := &Node{
		config:      config,
		randomizer:  Tools.NewRandomizer(config.RandomizerSeed),
		application: application,
	}
	if node.config.MailerConfig != nil {
		node.mailer = Tools.NewMailer(node.config.MailerConfig)
	}
	if node.config.InfoLoggerPath != "" {
		node.infoLogger = Tools.NewLogger("[InternalInfo: \""+node.config.Name+"\"] ", node.config.InfoLoggerPath)
	}
	if node.config.WarningLoggerPath != "" {
		node.warningLogger = Tools.NewLogger("[InternalWarning: \""+node.config.Name+"\"] ", node.config.WarningLoggerPath)
	}
	if node.config.ErrorLoggerPath != "" {
		node.errorLogger = Tools.NewLogger("[Error: \""+node.config.Name+"\"] ", node.config.ErrorLoggerPath)
	}
	return node
}

// StartBlocking starts the node and blocks until the node is stopped.
func (node *Node) StartBlocking() error {
	if node.status != STATUS_STOPPED {
		return Error.New("node already started", nil)
	}
	err := node.Start()
	<-node.stopChannel
	return err
}

// Start starts the node.
// Blocks until the node is fully started.
func (node *Node) Start() error {
	node.startMutex.Lock()
	defer node.startMutex.Unlock()
	if node.status != STATUS_STOPPED {
		return Error.New("node not stopped", nil)
	}
	node.status = STATUS_STARTING

	if infoLogger := node.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Starting", nil).Error())
	}
	node.stopChannel = make(chan bool)

	if ImplementsSystemgeServerComponent(node.application) {
		if infoLogger := node.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Starting systemge component", nil).Error())
		}
		err := node.startSystemgeServerComponent()
		if err != nil {
			if err := node.Stop(); err != nil {
				if warningLogger := node.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("failed to stop node", err).Error())
				}
			}
			return Error.New("failed starting systemge component", err)
		}
		if infoLogger := node.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Started systemge component", nil).Error())
		}
	}

	if node.config.SystemgeClientConfig != nil {
		if infoLogger := node.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Starting systemge client component", nil).Error())
		}
		err := node.startSystemgeClientComponent()
		if err != nil {
			if err := node.Stop(); err != nil {
				if warningLogger := node.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("failed to stop node", err).Error())
				}
			}
			return Error.New("failed starting systemge client component", err)
		}
		if infoLogger := node.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Started systemge client component", nil).Error())
		}
	}

	if ImplementsOnStartComponent(node.application) {
		if infoLogger := node.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Executing OnStart", nil).Error())
		}
		err := node.GetOnStartComponent().OnStart(node)
		if err != nil {
			if err := node.Stop(); err != nil {
				node.warningLogger.Log(Error.New("failed to stop node. Error in OnStart", err).Error())
			}
			return Error.New("failed in OnStart", err)
		}
		if infoLogger := node.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("executed OnStart", nil).Error())
		}
	}
	if infoLogger := node.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Started", nil).Error())
	}
	node.status = STATUS_STARTED
	return nil
}

// Stop stops the node.
// Blocks until the node is fully stopped.
func (node *Node) Stop() error {
	node.stopMutex.Lock()
	defer node.stopMutex.Unlock()
	if node.status == STATUS_STOPPED {
		return Error.New("node already stopped", nil)
	}
	if node.status == STATUS_STOPPING {
		return Error.New("node already stopping", nil)
	}
	status := node.status
	node.status = STATUS_STOPPING

	if infoLogger := node.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Stopping", nil).Error())
	}

	if status == STATUS_STARTED {
		if ImplementsOnStopComponent(node.application) {
			if infoLogger := node.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Executing OnStop", nil).Error())
			}
			err := node.GetOnStopComponent().OnStop(node)
			if err != nil {
				node.status = STATUS_STARTED
				return Error.New("failed to stop node. Error in OnStop", err)
			}
			if infoLogger := node.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("executed OnStop", nil).Error())
			}
		}
	}

	if node.systemgeClient != nil {
		if infoLogger := node.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Stopping systemge client component", nil).Error())
		}
		node.stopSystemgeClientComponent()
		if infoLogger := node.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Stopped systemge client component", nil).Error())
		}
	}

	if node.systemgeServer != nil {
		if infoLogger := node.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Stopping systemge component", nil).Error())
		}
		node.stopSystemgeServerComponent()
		if infoLogger := node.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Stopped systemge component", nil).Error())
		}
	}
	close(node.stopChannel)
	if infoLogger := node.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Stopped", nil).Error())
	}
	node.status = STATUS_STOPPED
	return nil
}

func (node *Node) GetStatus() int {
	return node.status
}

func (node *Node) GetName() string {
	return node.config.Name
}

func (node *Node) GetApplication() Application {
	return node.application
}

func (node *Node) GetRandomizer() *Tools.Randomizer {
	return node.randomizer
}

func (node *Node) SetRandomizer(randomizer *Tools.Randomizer) {
	node.randomizer = randomizer
}

func (node *Node) GetSystemgeEndpointConfig() *Config.TcpEndpoint {
	if node.config.SystemgeServerConfig == nil {
		return nil
	}
	return node.config.SystemgeServerConfig.Endpoint
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
