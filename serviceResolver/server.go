package serviceResolver

import (
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceSingleRequest"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type Resolver[B any, O any] struct {
	topicObjects map[string]O
	mutex        sync.RWMutex

	singleRequestServerSync *serviceSingleRequest.SingleRequestServerSync[B]

	deserializeTopic func(B, systemge.Connection[B]) (string, error)

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
		topicObjects:     make(map[string]O),
		deserializeTopic: deserializeTopic,
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

	singleRequestServerSync, err := serviceSingleRequest.NewSingleRequestServerSync(singleRequestServerConfig, routineConfig, listener, acceptHandler, readHandlerWrapper)
	if err != nil {
		return nil, err
	}
	resolver.singleRequestServerSync = singleRequestServerSync

	return resolver, nil
}

func (resolver *Resolver[B, O]) GetSingleRequestServerSync() *serviceSingleRequest.SingleRequestServerSync[B] {
	return resolver.singleRequestServerSync
}

func (resolver *Resolver[B, O]) AddResolution(topic string, resolution O) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	resolver.topicObjects[topic] = resolution
}

func (resolver *Resolver[B, O]) RemnoveResolution(topic string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	delete(resolver.topicObjects, topic)
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

func (server *Resolver[B, O]) GetDefaultCommands() tools.CommandHandlers {
	commands := tools.CommandHandlers{
		"start": func(args []string) (string, error) {
			err := server.singleRequestServerSync.GetRoutine().StartRoutine()
			if err != nil {
				return "", err
			}
			return "success", nil
		},
		"stop": func(args []string) (string, error) {
			err := server.singleRequestServerSync.GetRoutine().StopRoutine(true)
			if err != nil {
				return "", err
			}
			return "success", nil
		},
		"getStatus": func(args []string) (string, error) {
			return status.ToString(server.singleRequestServerSync.GetRoutine().GetStatus()), nil
		},
		"checkMetrics": func(args []string) (string, error) {
			metrics := server.CheckMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", err
			}
			return string(json), nil
		},
		"getMetrics": func(args []string) (string, error) {
			metrics := server.GetMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", err
			}
			return string(json), nil
		},
		/* "addResolution": func(args []string) (string, error) {
			if len(args) != 2 {
				return "", nil
			}
			tcpClientConfig := configs.UnmarshalTcpClient(args[1])
			if tcpClientConfig == nil {
				return "", nil
			}
			server.AddResolution(args[0], tcpClientConfig)
			return "success", nil
		}, */
		"removeResolution": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", errors.New("expected 1 argument")
			}
			server.RemnoveResolution(args[0])
			return "success", nil
		},
	}
	systemgeServerCommands := server.singleRequestServerSync.GetDefaultCommands()
	for key, value := range systemgeServerCommands {
		commands["systemgeServer_"+key] = value
	}
	return commands
}
