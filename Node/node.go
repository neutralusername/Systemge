package Node

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tools"
)

type Node struct {
	config        *Config.Node
	newNodeConfig *Config.NewNode

	stopChannel chan bool
	status      int
	mutex       sync.Mutex

	randomizer            *Tools.Randomizer
	mailer                *Tools.Mailer
	infoLogger            *Tools.Logger
	internalInfoLogger    *Tools.Logger
	warningLogger         *Tools.Logger
	internalWarningLogger *Tools.Logger
	errorLogger           *Tools.Logger
	debugLogger           *Tools.Logger

	application Application
	systemge    *systemgeComponent
	websocket   *websocketComponent
	http        *httpComponent
}

const (
	STATUS_STOPPED  = 0
	STATUS_STARTING = 1
	STATUS_STARTED  = 2
	STATUS_STOPPING = 3
)

func New(config *Config.NewNode, application Application) *Node {
	if config == nil {
		panic("NewNodeConfig is required")
	}
	if config.NodeConfig == nil {
		panic("NodeConfig is required")
	}
	if application == nil {
		panic("Application is required")
	}
	if config.HttpConfig == nil && config.WebsocketConfig == nil && config.SystemgeConfig == nil {
		panic("At least one of the following configurations is required: HttpConfig, WebsocketConfig, SystemgeConfig")
	}
	implementsHTTPComponent := ImplementsHTTPComponent(application)
	implementsWebsocketComponent := ImplementsWebsocketComponent(application)
	implementsSystemgeComponent := ImplementsSystemgeComponent(application)
	if !implementsHTTPComponent && !implementsWebsocketComponent && !implementsSystemgeComponent {
		panic("Application must implement at least one of the following interfaces: HTTPComponent, WebsocketComponent, SystemgeComponent")
	}
	if !implementsHTTPComponent && config.HttpConfig != nil {
		panic("HttpConfig provided but application does not implement HTTPComponent")
	}
	if !implementsWebsocketComponent && config.WebsocketConfig != nil {
		panic("WebsocketConfig provided but application does not implement WebsocketComponent")
	}
	if !implementsSystemgeComponent && config.SystemgeConfig != nil {
		panic("SystemgeConfig provided but application does not implement SystemgeComponent")
	}
	if implementsHTTPComponent && config.HttpConfig == nil {
		panic("Application implements HTTPComponent but HttpConfig is missing")
	}
	if implementsWebsocketComponent && config.WebsocketConfig == nil {
		panic("Application implements WebsocketComponent but WebsocketConfig is missing")
	}
	if implementsSystemgeComponent && config.SystemgeConfig == nil {
		panic("Application implements SystemgeComponent but SystemgeConfig is missing")
	}
	if implementsSystemgeComponent && config.SystemgeConfig.ServerConfig == nil {
		panic("Application implements SystemgeComponent but SystemgeConfig.ServerConfig is missing")
	}
	if implementsWebsocketComponent && config.WebsocketConfig.ServerConfig == nil {
		panic("Application implements WebsocketComponent but WebsocketConfig.ServerConfig is missing")
	}
	if implementsWebsocketComponent && config.WebsocketConfig.Pattern == "" {
		panic("Application implements WebsocketComponent but WebsocketConfig.Pattern is missing")
	}
	if implementsHTTPComponent && config.HttpConfig.ServerConfig == nil {
		panic("Application implements HTTPComponent but HttpConfig.ServerConfig is missing")
	}
	node := &Node{
		newNodeConfig: config,
		config:        config.NodeConfig,
		randomizer:    Tools.NewRandomizer(config.NodeConfig.RandomizerSeed),

		application: application,
	}
	if node.config.MailerConfig != nil {
		node.mailer = Tools.NewMailer(node.config.MailerConfig)
	}
	if node.config.InfoLoggerPath != "" {
		node.infoLogger = Tools.NewLogger("[Info: \""+node.config.Name+"\"] ", node.config.InfoLoggerPath)
	}
	if node.config.InternalInfoLoggerPath != "" {
		node.internalInfoLogger = Tools.NewLogger("[InternalInfo: \""+node.config.Name+"\"] ", node.config.InternalInfoLoggerPath)
	}
	if node.config.WarningLoggerPath != "" {
		node.warningLogger = Tools.NewLogger("[Warning: \""+node.config.Name+"\"] ", node.config.WarningLoggerPath)
	}
	if node.config.InternalWarningLoggerPath != "" {
		node.internalWarningLogger = Tools.NewLogger("[InternalWarning: \""+node.config.Name+"\"] ", node.config.InternalWarningLoggerPath)
	}
	if node.config.ErrorLoggerPath != "" {
		node.errorLogger = Tools.NewLogger("[Error: \""+node.config.Name+"\"] ", node.config.ErrorLoggerPath)
	}
	if node.config.DebugLoggerPath != "" {
		node.debugLogger = Tools.NewLogger("[Debug: \""+node.config.Name+"\"] ", node.config.DebugLoggerPath)
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
	node.mutex.Lock()
	if node.status != STATUS_STOPPED {
		node.mutex.Unlock()
		return Error.New("node not stopped", nil)
	}
	node.status = STATUS_STARTING
	node.mutex.Unlock()

	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Starting", nil).Error())
	}
	node.stopChannel = make(chan bool)
	if ImplementsSystemgeComponent(node.application) {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Starting systemge component", nil).Error())
		}
		err := node.startSystemgeComponent()
		if err != nil {
			if err := node.Stop(); err != nil {
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("failed to stop node", err).Error())
				}
			}
			return Error.New("failed starting systemge component", err)
		}
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Started systemge component", nil).Error())
		}
	}
	if ImplementsWebsocketComponent(node.application) {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Starting websocket component", nil).Error())
		}
		err := node.startWebsocketComponent()
		if err != nil {
			if err := node.Stop(); err != nil {
				node.GetInternalWarningError().Log(Error.New("failed to stop node. Error in OnStart", err).Error())
			}
			return Error.New("failed starting websocket server", err)
		}
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Started websocket component", nil).Error())
		}
	}
	if ImplementsHTTPComponent(node.application) {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Starting http component", nil).Error())
		}
		err := node.startHTTPComponent()
		if err != nil {
			if err := node.Stop(); err != nil {
				node.GetInternalWarningError().Log(Error.New("failed to stop node. Error in OnStart", err).Error())
			}
			return Error.New("failed starting http server", err)
		}
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Started http component", nil).Error())
		}
	}
	if ImplementsOnStartComponent(node.application) {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Executing OnStart", nil).Error())
		}
		err := node.GetOnStartComponent().OnStart(node)
		if err != nil {
			if err := node.Stop(); err != nil {
				node.GetInternalWarningError().Log(Error.New("failed to stop node. Error in OnStart", err).Error())
			}
			return Error.New("failed in OnStart", err)
		}
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("executed OnStart", nil).Error())
		}
	}
	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Started", nil).Error())
	}
	node.status = STATUS_STARTED
	return nil
}

// Stop stops the node.
// Blocks until the node is fully stopped.
func (node *Node) Stop() error {
	node.mutex.Lock()
	if node.status == STATUS_STOPPED {
		node.mutex.Unlock()
		return Error.New("node already stopped", nil)
	}
	if node.status == STATUS_STOPPING {
		node.mutex.Unlock()
		return Error.New("node already stopping", nil)
	}
	status := node.status
	node.status = STATUS_STOPPING
	node.mutex.Unlock()

	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Stopping", nil).Error())
	}
	if status == STATUS_STARTED {
		if ImplementsOnStopComponent(node.application) {
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Executing OnStop", nil).Error())
			}
			err := node.GetOnStopComponent().OnStop(node)
			if err != nil {
				node.status = STATUS_STARTED
				return Error.New("failed to stop node. Error in OnStop", err)
			}
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("executed OnStop", nil).Error())
			}
		}
	}
	if node.websocket != nil {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopping websocket component", nil).Error())
		}
		node.stopWebsocketComponent()
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopped websocket component", nil).Error())
		}
	}
	if node.http != nil {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopping http component", nil).Error())
		}
		node.stopHTTPComponent()
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopped http component", nil).Error())
		}
	}
	if node.systemge != nil {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopping systemge component", nil).Error())
		}
		node.stopSystemgeComponent()
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopped systemge component", nil).Error())
		}
	}
	close(node.stopChannel)
	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
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

func (node *Node) GetErrorLogger() *Tools.Logger {
	return node.errorLogger
}

func (node *Node) GetInternalWarningError() *Tools.Logger {
	return node.internalWarningLogger
}

func (node *Node) GetWarningLogger() *Tools.Logger {
	return node.warningLogger
}

func (node *Node) GetInfoLogger() *Tools.Logger {
	return node.infoLogger
}

func (node *Node) GetInternalInfoLogger() *Tools.Logger {
	return node.internalInfoLogger
}

func (node *Node) GetDebugLogger() *Tools.Logger {
	return node.debugLogger
}

func (node *Node) GetMailer() *Tools.Mailer {
	return node.mailer
}

func (node *Node) GetRandomizer() *Tools.Randomizer {
	return node.randomizer
}

func (node *Node) SetRandomizer(randomizer *Tools.Randomizer) {
	node.randomizer = randomizer
}

func (node *Node) GetSystemgeEndpointConfig() *Config.TcpEndpoint {
	if node.newNodeConfig.SystemgeConfig == nil {
		return nil
	}
	return node.newNodeConfig.SystemgeConfig.Endpoint
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
