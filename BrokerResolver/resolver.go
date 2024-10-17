package BrokerResolver

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/singleRequestServer"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type Resolver[B any] struct {
	name string

	config *configs.MessageBrokerResolver

	topicTcpClientConfigs map[string]*configs.TcpClient
	mutex                 sync.Mutex

	ongoingResolutions atomic.Int64

	singleRequestServer *singleRequestServer.SingleRequestServerSync[B]

	// metrics

	SucessfulResolutions atomic.Uint64
	FailedResolutions    atomic.Uint64
}

func New[B any](
	topicClientConfigs map[string]*configs.TcpClient,
	singleRequestServerConfig *configs.SingleRequestServerSync,
	routineConfig *configs.Routine,
	listener systemge.Listener[B, systemge.Connection[B]],
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]],
	readHandler tools.ReadHandlerWithError[B, systemge.Connection[B]],
	deserializeTopic func(B) (string, error),
	serializeResolution func(*configs.TcpClient) (B, error),
) (*Resolver[B], error) {

	resolver := &Resolver[B]{
		topicTcpClientConfigs: make(map[string]*configs.TcpClient),
	}

	readHandlerWrapper := func(data B, connection systemge.Connection[B]) (B, error) {
		if err := readHandler(data, connection); err != nil {
			var nilValue B
			return nilValue, err
		}
		topic, err := deserializeTopic(data)
		if err != nil {
			var nilValue B
			return nilValue, err
		}

		resolver.mutex.Lock()
		tcpClientConfig, ok := topicClientConfigs[topic]
		resolver.mutex.Unlock()

		if !ok {
			var nilValue B
			return nilValue, errors.New("topic not found")
		}
		return serializeResolution(tcpClientConfig)
	}

	singleRequestServerSync, err := singleRequestServer.NewSingleRequestServerSync(singleRequestServerConfig, routineConfig, listener, acceptHandler, readHandlerWrapper)
	if err != nil {
		return nil, err
	}
	resolver.singleRequestServer = singleRequestServerSync

	for topic, tcpClientConfig := range topicClientConfigs {
		normalizedAddress, err := helpers.NormalizeAddress(tcpClientConfig.Address)
		if err != nil {
			return nil, err
		}
		tcpClientConfig.Address = normalizedAddress
		resolver.topicTcpClientConfigs[topic] = tcpClientConfig
	}

	return resolver, nil
}
