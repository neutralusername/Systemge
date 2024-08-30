package MessageBroker

import (
	"sync"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Dashboard"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
	"github.com/neutralusername/Systemge/Tools"
)

type Resolver struct {
	config *Config.MessageBrokerResolver

	systemgeServer *SystemgeServer.SystemgeServer

	dashboardClient *Dashboard.DashboardClient

	asyncTopicResolutions map[string]*Config.TcpEndpoint
	syncTopicResolutions  map[string]*Config.TcpEndpoint
	mutex                 sync.Mutex

	messageHandler SystemgeConnection.MessageHandler

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer
}

func NewResolver(config *Config.MessageBrokerResolver) *Resolver {
	if config == nil {
		panic("Config is required")
	}
	if config.SystemgeServerConfig == nil {
		panic("SystemgeServerConfig is required")
	}
	if config.SystemgeServerConfig.ConnectionConfig == nil {
		panic("SystemgeServerConfig.ConnectionConfig is required")
	}
	if config.SystemgeServerConfig.Name == "" {
		panic("SystemgeServerConfig.Name is required")
	}
	if config.SystemgeServerConfig.ConnectionConfig.TcpBufferBytes == 0 {
		config.SystemgeServerConfig.ConnectionConfig.TcpBufferBytes = 1024 * 4
	}

	resolver := &Resolver{
		config:                config,
		asyncTopicResolutions: make(map[string]*Config.TcpEndpoint),
		syncTopicResolutions:  make(map[string]*Config.TcpEndpoint),
	}

	if config.InfoLoggerPath != "" {
		resolver.infoLogger = Tools.NewLogger("[Info: \""+config.Name+"\"]", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		resolver.warningLogger = Tools.NewLogger("[Warning: \""+config.Name+"\"]", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		resolver.errorLogger = Tools.NewLogger("[Error: \""+config.Name+"\"]", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		resolver.mailer = Tools.NewMailer(config.MailerConfig)
	}

	for topic, resolution := range config.AsyncTopicResolutions {
		resolver.asyncTopicResolutions[topic] = resolution
	}
	for topic, resolution := range config.SyncTopicResolutions {
		resolver.syncTopicResolutions[topic] = resolution
	}

	resolver.messageHandler = SystemgeConnection.NewConcurrentMessageHandler(nil, SystemgeConnection.SyncMessageHandlers{
		Message.TOPIC_RESOLVE_ASYNC: resolver.resolveAsync,
		Message.TOPIC_RESOLVE_SYNC:  resolver.resolveSync,
	}, nil, nil)

	resolver.systemgeServer = SystemgeServer.New(config.SystemgeServerConfig, resolver.onConnect, nil)

	if config.DashboardClientConfig != nil {
		resolver.dashboardClient = Dashboard.NewClient(config.DashboardClientConfig, resolver.systemgeServer.Start, resolver.systemgeServer.Stop, resolver.GetMetrics, resolver.systemgeServer.GetStatus, Commands.Handlers{
			"add_async_resolution": func(args []string) (string, error) {
				if len(args) != 2 {
					return "", Error.New("Invalid number of arguments (expected 1)", nil)
				}
				endpoint := Config.UnmarshalTcpEndpoint(args[1])
				if endpoint == nil {
					return "", Error.New("Invalid endpoint in json format provided", nil)
				}
				resolver.AddAsyncResolution(args[0], endpoint)
				return "Success", nil
			},
			"add_sync_resolution": func(args []string) (string, error) {
				if len(args) != 2 {
					return "", Error.New("Invalid number of arguments (expected 1)", nil)
				}
				endpoint := Config.UnmarshalTcpEndpoint(args[1])
				if endpoint == nil {
					return "", Error.New("Invalid endpoint in json format provided", nil)
				}
				resolver.AddSyncResolution(args[0], endpoint)
				return "Success", nil
			},
			"remove_async_resolution": func(args []string) (string, error) {
				if len(args) != 1 {
					return "", Error.New("Invalid number of arguments (expected 1)", nil)
				}
				resolver.RemoveAsyncResolution(args[0])
				return "Success", nil
			},
			"remove_sync_resolution": func(args []string) (string, error) {
				if len(args) != 1 {
					return "", Error.New("Invalid number of arguments (expected 1)", nil)
				}
				resolver.RemoveSyncResolution(args[0])
				return "Success", nil
			},
		})
		if err := resolver.StartDashboardClient(); err != nil {
			if resolver.errorLogger != nil {
				resolver.errorLogger.Log(Error.New("Failed to start dashboard client", err).Error())
			}
			if resolver.mailer != nil {
				if err := resolver.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to start dashboard client", err).Error())); err != nil {
					if resolver.errorLogger != nil {
						resolver.errorLogger.Log(Error.New("Failed to send email", err).Error())
					}
				}
			}
		}
	}

	return resolver
}

func (resolver *Resolver) Start() error {
	return resolver.systemgeServer.Start()
}

func (resolver *Resolver) Stop() error {
	return resolver.systemgeServer.Stop()
}

func (resolver *Resolver) GetStatus() int {
	return resolver.systemgeServer.GetStatus()
}

func (resolver *Resolver) GetMetrics() map[string]uint64 {
	// TODO: gather metrics
	return nil
}

func (resolver *Resolver) StartDashboardClient() error {
	if resolver.dashboardClient == nil {
		return Error.New("Dashboard client is not configured", nil)
	}
	return resolver.dashboardClient.Start()
}

func (resolver *Resolver) StopDashboardClient() error {
	if resolver.dashboardClient == nil {
		return Error.New("Dashboard client is not configured", nil)
	}
	return resolver.dashboardClient.Stop()
}

func (resolver *Resolver) AddAsyncResolution(topic string, resolution *Config.TcpEndpoint) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	resolver.asyncTopicResolutions[topic] = resolution
}

func (resolver *Resolver) AddSyncResolution(topic string, resolution *Config.TcpEndpoint) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	resolver.syncTopicResolutions[topic] = resolution
}

func (resolver *Resolver) RemoveAsyncResolution(topic string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	delete(resolver.asyncTopicResolutions, topic)
}

func (resolver *Resolver) RemoveSyncResolution(topic string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	delete(resolver.syncTopicResolutions, topic)
}

func (resolver *Resolver) resolveAsync(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if resolution, ok := resolver.asyncTopicResolutions[message.GetTopic()]; ok {
		return Helpers.JsonMarshal(resolution), nil
	} else {
		return "", Error.New("Unkown topic", nil)
	}
}

func (resolver *Resolver) resolveSync(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if resolution, ok := resolver.syncTopicResolutions[message.GetTopic()]; ok {
		return Helpers.JsonMarshal(resolution), nil
	} else {
		return "", Error.New("Unkown topic", nil)
	}
}

func (resolver *Resolver) onConnect(connection *SystemgeConnection.SystemgeConnection) error {
	message, err := connection.GetNextMessage()
	if err != nil {
		return err
	}
	switch message.GetTopic() {
	case Message.TOPIC_RESOLVE_ASYNC:
		fallthrough
	case Message.TOPIC_RESOLVE_SYNC:
		err := connection.ProcessMessage(message, resolver.messageHandler)
		if err != nil {
			return err
		}
		connection.Close()
		return nil
	default:
		return Error.New("Invalid topic", nil)
	}
}
