package resolverServer

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/singleRequestServer"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type Resolver[B any, O any] struct {
	config *configs.MessageBrokerResolver

	topicObjects map[string]O
	mutex        sync.RWMutex

	singleRequestServerSync *singleRequestServer.SingleRequestServerSync[B]

	// metrics

	SucessfulResolutions atomic.Uint64
	FailedResolutions    atomic.Uint64
}

func New[B any, O any](
	topicObject map[string]O,
	singleRequestServerConfig *configs.SingleRequestServerSync,
	routineConfig *configs.Routine,
	listener systemge.Listener[B, systemge.Connection[B]],
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]],
	deserializeTopic func(B, systemge.Connection[B]) (string, error), // responsible for validating the request and retrieving the topic
	serializeObject func(O) (B, error),
) (*Resolver[B, O], error) {

	resolver := &Resolver[B, O]{
		topicObjects: make(map[string]O),
	}

	for topic, object := range topicObject {
		resolver.topicObjects[topic] = object
	}

	readHandlerWrapper := func(data B, connection systemge.Connection[B]) (B, error) {
		topic, err := deserializeTopic(data, connection)
		if err != nil {
			var nilValue B
			return nilValue, err
		}

		resolver.mutex.RLock()
		object, ok := topicObject[topic]
		resolver.mutex.RUnlock()

		if !ok {
			var nilValue B
			return nilValue, errors.New("topic not found")
		}
		return serializeObject(object)
	}

	singleRequestServerSync, err := singleRequestServer.NewSingleRequestServerSync(singleRequestServerConfig, routineConfig, listener, acceptHandler, readHandlerWrapper)
	if err != nil {
		return nil, err
	}
	resolver.singleRequestServerSync = singleRequestServerSync

	return resolver, nil
}

func (resolver *Resolver[B, O]) GetSingleRequestServerSync() *singleRequestServer.SingleRequestServerSync[B] {
	return resolver.singleRequestServerSync
}

func (resolver *Resolver[B, O]) CheckMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	resolver.mutex.RLock()
	metricsTypes.AddMetrics("brokerResolver_resolutions", tools.NewMetrics(
		map[string]uint64{
			"successes": resolver.SucessfulResolutions.Load(),
			"failures":  resolver.FailedResolutions.Load(),
			"topics":    uint64(len(resolver.topicObjects)),
		},
	))
	resolver.mutex.RUnlock()
	metricsTypes.Merge(resolver.singleRequestServerSync.CheckMetrics())
	return metricsTypes
}
func (resolver *Resolver[B, O]) GetMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	resolver.mutex.RLock()
	metricsTypes.AddMetrics("brokerResolver", tools.NewMetrics(
		map[string]uint64{
			"successes": resolver.SucessfulResolutions.Swap(0),
			"failures":  resolver.FailedResolutions.Swap(0),
			"topics":    uint64(len(resolver.topicObjects)),
		},
	))
	resolver.mutex.RUnlock()
	metricsTypes.Merge(resolver.singleRequestServerSync.GetMetrics())
	return metricsTypes
}
