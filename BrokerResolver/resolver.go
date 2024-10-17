package BrokerResolver

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type Resolver[B any] struct {
	name string

	config *configs.MessageBrokerResolver

	asyncTopicTcpClientConfigs map[string]*configs.TcpClient
	syncTopicTcpClientConfigs  map[string]*configs.TcpClient
	mutex                      sync.Mutex

	ongoingResolutions atomic.Int64

	// metrics

	SucessfulAsyncResolutions atomic.Uint64
	SucessfulSyncResolutions  atomic.Uint64
	FailedResolutions         atomic.Uint64
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

	return resolver, nil
}
