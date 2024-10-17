package BrokerResolver

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/systemge/Config"
	"github.com/neutralusername/systemge/Message"
	"github.com/neutralusername/systemge/Server"
	"github.com/neutralusername/systemge/SystemgeConnection"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/tools"
)

type Resolver struct {
	name string

	config *Config.MessageBrokerResolver

	systemgeServer *Server.Server

	asyncTopicTcpClientConfigs map[string]*Config.TcpClient
	syncTopicTcpClientConfigs  map[string]*Config.TcpClient
	mutex                      sync.Mutex

	messageHandler SystemgeConnection.MessageHandler

	infoLogger    *tools.Logger
	warningLogger *tools.Logger
	errorLogger   *tools.Logger
	mailer        *tools.Mailer

	ongoingResolutions atomic.Int64

	// metrics

	sucessfulAsyncResolutions atomic.Uint64
	sucessfulSyncResolutions  atomic.Uint64
	failedResolutions         atomic.Uint64
}

func New(name string, config *Config.MessageBrokerResolver, whitelist *tools.AccessControlList, blacklist *tools.AccessControlList) *Resolver {
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
		resolver.infoLogger = tools.NewLogger("[Info: \""+name+"\"]", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		resolver.warningLogger = tools.NewLogger("[Warning: \""+name+"\"]", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		resolver.errorLogger = tools.NewLogger("[Error: \""+name+"\"]", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		resolver.mailer = tools.NewMailer(config.MailerConfig)
	}

	for topic, tcpClientConfig := range config.AsyncTopicClientConfigs {
		normalizedAddress, err := helpers.NormalizeAddress(tcpClientConfig.Address)
		if err != nil {
			panic(err)
		}
		tcpClientConfig.Address = normalizedAddress
		resolver.asyncTopicTcpClientConfigs[topic] = tcpClientConfig
	}
	for topic, tcpClientConfig := range config.SyncTopicClientConfigs {
		normalizedAddress, err := helpers.NormalizeAddress(tcpClientConfig.Address)
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

	resolver.systemgeServer = Server.New(name+"_systemgeServer",
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
