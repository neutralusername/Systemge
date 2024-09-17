package BrokerResolver

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
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

	asyncTopicTcpClientConfigs map[string]*Config.TcpClient
	syncTopicTcpClientConfigs  map[string]*Config.TcpClient
	mutex                      sync.Mutex

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
	if config.SystemgeServerConfig.TcpSystemgeConnectionConfig == nil {
		panic("SystemgeServerConfig.ConnectionConfig is required")
	}
	if name == "" {
		panic("SystemgeServerConfig.Name is required")
	}

	resolver := &Resolver{
		name:                       name,
		config:                     config,
		asyncTopicTcpClientConfigs: make(map[string]*Config.TcpClient),
		syncTopicTcpClientConfigs:  make(map[string]*Config.TcpClient),
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

	for topic, tcpClientConfig := range config.AsyncTopicClientConfigs {
		normalizedAddress, err := Helpers.NormalizeAddress(tcpClientConfig.Address)
		if err != nil {
			panic(err)
		}
		tcpClientConfig.Address = normalizedAddress
		resolver.asyncTopicTcpClientConfigs[topic] = tcpClientConfig
	}
	for topic, tcpClientConfig := range config.SyncTopicClientConfigs {
		normalizedAddress, err := Helpers.NormalizeAddress(tcpClientConfig.Address)
		if err != nil {
			panic(err)
		}
		tcpClientConfig.Address = normalizedAddress
		resolver.syncTopicTcpClientConfigs[topic] = tcpClientConfig
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

func (resolver *Resolver) GetName() string {
	return resolver.name
}
