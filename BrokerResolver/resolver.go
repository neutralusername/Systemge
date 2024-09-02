package BrokerResolver

import (
	"sync"
	"sync/atomic"

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
	name string

	config *Config.MessageBrokerResolver

	systemgeServer *SystemgeServer.SystemgeServer

	dashboardClient *Dashboard.DashboardClient

	asyncTopicEndpoints map[string]*Config.TcpClient
	syncTopicEndpoints  map[string]*Config.TcpClient
	mutex               sync.Mutex

	messageHandler SystemgeConnection.MessageHandler

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer

	// metrics

	sucessfulAsyncResolutions atomic.Uint64
	sucessfulSyncResolutions  atomic.Uint64
	failedResolutions         atomic.Uint64
}

func New(name string, config *Config.MessageBrokerResolver) *Resolver {
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

	for topic, endpoint := range config.AsyncTopicEndpoints {
		normalizedAddress, err := Helpers.NormalizeAddress(endpoint.Address)
		if err != nil {
			panic(err)
		}
		endpoint.Address = normalizedAddress
		resolver.asyncTopicEndpoints[topic] = endpoint
	}
	for topic, endpoint := range config.SyncTopicEndpoints {
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

	resolver.systemgeServer = SystemgeServer.New(name+"_systemgeServer", config.SystemgeServerConfig, resolver.onConnect, nil)

	if config.DashboardClientConfig != nil {
		resolver.dashboardClient = Dashboard.NewClient(config.DashboardClientConfig, resolver.systemgeServer.Start, resolver.systemgeServer.Stop, resolver.GetMetrics, resolver.systemgeServer.GetStatus, Commands.Handlers{
			"add_async_resolution": func(args []string) (string, error) {
				if len(args) != 2 {
					return "", Error.New("Invalid number of arguments (expected 1)", nil)
				}
				endpoint := Config.UnmarshalTcpClient(args[1])
				if endpoint == nil {
					return "", Error.New("Invalid endpoint in json format provided", nil)
				}
				normalizedAddress, err := Helpers.NormalizeAddress(endpoint.Address)
				if err != nil {
					return "", err
				}
				endpoint.Address = normalizedAddress
				resolver.AddAsyncResolution(args[0], endpoint)
				return "Success", nil
			},
			"add_sync_resolution": func(args []string) (string, error) {
				if len(args) != 2 {
					return "", Error.New("Invalid number of arguments (expected 1)", nil)
				}
				endpoint := Config.UnmarshalTcpClient(args[1])
				if endpoint == nil {
					return "", Error.New("Invalid endpoint in json format provided", nil)
				}
				normalizedAddress, err := Helpers.NormalizeAddress(endpoint.Address)
				if err != nil {
					return "", err
				}
				endpoint.Address = normalizedAddress
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

func (resolver *Resolver) GetMetrics() map[string]uint64 {
	metrics := resolver.systemgeServer.RetrieveMetrics()
	metrics["sucessful_async_resolutions"] = resolver.RetrieveSucessfulAsyncResolutions()
	metrics["sucessful_sync_resolutions"] = resolver.RetrieveSucessfulSyncResolutions()
	metrics["failed_resolutions"] = resolver.RetrieveFailedResolutions()
	return metrics
}
