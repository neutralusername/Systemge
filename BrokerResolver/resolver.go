package BrokerResolver

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/systemge/Message"
	"github.com/neutralusername/systemge/Server"
	"github.com/neutralusername/systemge/SystemgeConnection"
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type Resolver[B any] struct {
	name string

	config *configs.MessageBrokerResolver

	systemgeServer *Server.Server

	asyncTopicTcpClientConfigs map[string]*configs.TcpClient
	syncTopicTcpClientConfigs  map[string]*configs.TcpClient
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

func New[B any](
	asyncTopicClientConfigs map[string]*configs.TcpClient,
	syncTopicClientConfigs map[string]*configs.TcpClient,
	routineConfig *configs.Routine,
	listener systemge.Listener[B, systemge.Connection[B]],
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]],
	readHandler tools.ReadHandler[B, systemge.Connection[B]],
) (*Resolver[B], error) {

	resolver := &Resolver[B]{
		asyncTopicTcpClientConfigs: make(map[string]*configs.TcpClient),
		syncTopicTcpClientConfigs:  make(map[string]*configs.TcpClient),
	}

	for topic, tcpClientConfig := range asyncTopicClientConfigs {
		normalizedAddress, err := helpers.NormalizeAddress(tcpClientConfig.Address)
		if err != nil {
			return nil, err
		}
		tcpClientConfig.Address = normalizedAddress
		resolver.asyncTopicTcpClientConfigs[topic] = tcpClientConfig
	}
	for topic, tcpClientConfig := range syncTopicClientConfigs {
		normalizedAddress, err := helpers.NormalizeAddress(tcpClientConfig.Address)
		if err != nil {
			return nil, err
		}
		tcpClientConfig.Address = normalizedAddress
		resolver.syncTopicTcpClientConfigs[topic] = tcpClientConfig
	}

	resolver.messageHandler = SystemgeConnection.NewConcurrentMessageHandler(nil, SystemgeConnection.SyncMessageHandlers{
		Message.TOPIC_RESOLVE_ASYNC: resolver.resolveAsync,
		Message.TOPIC_RESOLVE_SYNC:  resolver.resolveSync,
	}, nil, nil)

	return resolver, nil
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
