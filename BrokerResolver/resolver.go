package BrokerResolver

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/singleRequestServer"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type Resolver[B any, T any] struct {
	config *configs.MessageBrokerResolver

	topicTcpClientConfigs map[string]T
	mutex                 sync.RWMutex

	singleRequestServerSync *singleRequestServer.SingleRequestServerSync[B]

	// metrics

	SucessfulResolutions atomic.Uint64
	FailedResolutions    atomic.Uint64
}

func New[B any, T any](
	topicClientConfigs map[string]T,
	singleRequestServerConfig *configs.SingleRequestServerSync,
	routineConfig *configs.Routine,
	listener systemge.Listener[B, systemge.Connection[B]],
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]],
	deserializeTopic func(B, systemge.Connection[B]) (string, error), // responsible for validating the request and retrieving the topic
	serializeResolution func(T) (B, error),
) (*Resolver[B, T], error) {

	resolver := &Resolver[B, T]{
		topicTcpClientConfigs: make(map[string]T),
	}

	for topic, tcpClientConfig := range topicClientConfigs {
		resolver.topicTcpClientConfigs[topic] = tcpClientConfig
	}

	readHandlerWrapper := func(data B, connection systemge.Connection[B]) (B, error) {
		topic, err := deserializeTopic(data, connection)
		if err != nil {
			var nilValue B
			return nilValue, err
		}

		resolver.mutex.RLock()
		tcpClientConfig, ok := topicClientConfigs[topic]
		resolver.mutex.RUnlock()

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
	resolver.singleRequestServerSync = singleRequestServerSync

	return resolver, nil
}

func (resolver *Resolver[B, T]) GetSingleRequestServerSync() *singleRequestServer.SingleRequestServerSync[B] {
	return resolver.singleRequestServerSync
}

func (resolver *Resolver[B, T]) CheckMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	resolver.mutex.RLock()
	metricsTypes.AddMetrics("brokerResolver_resolutions", tools.NewMetrics(
		map[string]uint64{
			"successes": resolver.SucessfulResolutions.Load(),
			"failures":  resolver.FailedResolutions.Load(),
			"topics":    uint64(len(resolver.topicTcpClientConfigs)),
		},
	))
	resolver.mutex.RUnlock()
	metricsTypes.Merge(resolver.singleRequestServerSync.CheckMetrics())
	return metricsTypes
}
func (resolver *Resolver[B, T]) GetMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	resolver.mutex.RLock()
	metricsTypes.AddMetrics("brokerResolver", tools.NewMetrics(
		map[string]uint64{
			"successes": resolver.SucessfulResolutions.Swap(0),
			"failures":  resolver.FailedResolutions.Swap(0),
			"topics":    uint64(len(resolver.topicTcpClientConfigs)),
		},
	))
	resolver.mutex.RUnlock()
	metricsTypes.Merge(resolver.singleRequestServerSync.GetMetrics())
	return metricsTypes
}
