package BrokerResolver

import (
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/DashboardClientCustom"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
	"github.com/neutralusername/Systemge/Tools"
)

type Resolver struct {
	name string

	config *Config.MessageBrokerResolver

	systemgeServer *SystemgeServer.SystemgeServer

	dashboardClient *DashboardClientCustom.Client

	asyncTopicEndpoints map[string]*Config.TcpClient
	syncTopicEndpoints  map[string]*Config.TcpClient
	mutex               sync.Mutex

	messageHandler SystemgeConnection.MessageHandler

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer

	ongoingResolutions atomic.Int64

	// metrics

	sucessfulAsyncResolutions atomic.Uint64
	sucessfulSyncResolutions  atomic.Uint64
	failedResolutions         atomic.Uint64
}

func New(name string, config *Config.MessageBrokerResolver, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList) *Resolver {
	if config == nil {
		panic("Config is required")
	}
	if config.SystemgeServerConfig == nil {
		panic("SystemgeServerConfig is required")
	}
	if config.SystemgeServerConfig.ConnectionConfig == nil {
		panic("SystemgeServerConfig.ConnectionConfig is required")
	}
	if name == "" {
		panic("SystemgeServerConfig.Name is required")
	}

	resolver := &Resolver{
		name:                name,
		config:              config,
		asyncTopicEndpoints: make(map[string]*Config.TcpClient),
		syncTopicEndpoints:  make(map[string]*Config.TcpClient),
	}

	if config.InfoLoggerPath != "" {
		resolver.infoLogger = Tools.NewLogger("[Info: \""+name+"\"]", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		resolver.warningLogger = Tools.NewLogger("[Warning: \""+name+"\"]", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		resolver.errorLogger = Tools.NewLogger("[Error: \""+name+"\"]", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		resolver.mailer = Tools.NewMailer(config.MailerConfig)
	}

	for topic, endpoint := range config.AsyncTopicClientConfigs {
		normalizedAddress, err := Helpers.NormalizeAddress(endpoint.Address)
		if err != nil {
			panic(err)
		}
		endpoint.Address = normalizedAddress
		resolver.asyncTopicEndpoints[topic] = endpoint
	}
	for topic, endpoint := range config.SyncTopicClientConfigs {
		normalizedAddress, err := Helpers.NormalizeAddress(endpoint.Address)
		if err != nil {
			panic(err)
		}
		endpoint.Address = normalizedAddress
		resolver.syncTopicEndpoints[topic] = endpoint
	}

	resolver.messageHandler = SystemgeConnection.NewConcurrentMessageHandler(nil, SystemgeConnection.SyncMessageHandlers{
		Message.TOPIC_RESOLVE_ASYNC: resolver.resolveAsync,
		Message.TOPIC_RESOLVE_SYNC:  resolver.resolveSync,
	}, nil, nil)

	resolver.systemgeServer = SystemgeServer.New(name+"_systemgeServer",
		config.SystemgeServerConfig,
		whitelist, blacklist,
		resolver.onConnect, nil,
	)

	if config.DashboardClientConfig != nil {
		resolver.dashboardClient = DashboardClientCustom.NewClient(name+"_dashboardClient",
			config.DashboardClientConfig,
			resolver.systemgeServer.Start, resolver.systemgeServer.Stop,
			resolver.GetMetrics, resolver.systemgeServer.GetStatus,
			resolver.GetDefaultCommands(),
		)
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

func (resolver *Resolver) GetStatus() int {
	return resolver.systemgeServer.GetStatus()
}

func (resolver *Resolver) GetName() string {
	return resolver.name
}

func (server *Resolver) GetDefaultCommands() Commands.Handlers {
	commands := Commands.Handlers{
		"start": func(args []string) (string, error) {
			err := server.Start()
			if err != nil {
				return "", Error.New("failed to start message broker server", err)
			}
			return "success", nil
		},
		"stop": func(args []string) (string, error) {
			err := server.Stop()
			if err != nil {
				return "", Error.New("failed to stop message broker server", err)
			}
			return "success", nil
		},
		"getStatus": func(args []string) (string, error) {
			return Status.ToString(server.GetStatus()), nil
		},
		"getMetrics": func(args []string) (string, error) {
			metrics := server.GetMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", Error.New("failed to marshal metrics to json", err)
			}
			return string(json), nil
		},
		"retrieveMetrics": func(args []string) (string, error) {
			metrics := server.RetrieveMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", Error.New("failed to marshal metrics to json", err)
			}
			return string(json), nil
		},
		"addAsyncResolution": func(args []string) (string, error) {
			if len(args) != 2 {
				return "", Error.New("expected 2 arguments", nil)
			}
			endpointConfig := Config.UnmarshalTcpClient(args[1])
			if endpointConfig == nil {
				return "", Error.New("failed unmarshalling endpointConfig", nil)
			}
			server.AddAsyncResolution(args[0], endpointConfig)
			return "success", nil
		},
		"removeAsyncResolution": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("expected 1 argument", nil)
			}
			server.RemoveAsyncResolution(args[0])
			return "success", nil
		},
		"addSyncResolution": func(args []string) (string, error) {
			if len(args) != 2 {
				return "", Error.New("expected 2 arguments", nil)
			}
			endpointConfig := Config.UnmarshalTcpClient(args[1])
			if endpointConfig == nil {
				return "", Error.New("failed unmarshalling endpointConfig", nil)
			}
			server.AddSyncResolution(args[0], endpointConfig)
			return "success", nil
		},
		"removeSyncResolution": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("expected 1 argument", nil)
			}
			server.RemoveSyncResolution(args[0])
			return "success", nil
		},
	}
	systemgeServerCommands := server.systemgeServer.GetDefaultCommands()
	for key, value := range systemgeServerCommands {
		commands["systemgeServer_"+key] = value
	}
	return commands
}
