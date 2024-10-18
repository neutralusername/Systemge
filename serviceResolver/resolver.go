package serviceResolver

import (
	"errors"
	"sync"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/serviceSingleRequest"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type Resolver[D any] struct {
	mutex     sync.RWMutex
	topicData map[string]D
	accepter  *serviceAccepter.Accepter[D]
	// commands
}

func New[D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterConfig *configs.AccepterServer,
	readerConfig *configs.ReaderServerSync,
	routineConfig *configs.Routine,
	topicData map[string]D,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	deserializeTopic func(D, systemge.Connection[D]) (string, error), // responsible for retrieving the topic
) (*Resolver[D], error) {

	resolver := &Resolver[D]{
		topicData: topicData,
	}
	accepter, err := serviceSingleRequest.NewSync(
		listener,
		accepterConfig,
		readerConfig,
		routineConfig,
		true,
		acceptHandler,
		func(closeChannel <-chan struct{}, incomingData D, connection systemge.Connection[D]) (D, error) {
			topic, err := deserializeTopic(incomingData, connection)
			if err != nil {
				return helpers.GetNilValue(incomingData), err
			}
			resolver.mutex.RLock()
			outgoingData, ok := topicData[topic]
			resolver.mutex.RUnlock()
			if !ok {
				return helpers.GetNilValue(incomingData), errors.New("topic not found")
			}
			return outgoingData, nil
		},
	)
	if err != nil {
		return nil, err
	}
	resolver.accepter = accepter
	return resolver, nil
}

func (r *Resolver[D]) GetAccepter() *serviceAccepter.Accepter[D] {
	return r.accepter
}

func (r *Resolver[D]) SetTopicData(topic string, data D) {
	r.mutex.Lock()
	r.topicData[topic] = data
	r.mutex.Unlock()
}

func (r *Resolver[D]) GetTopicData(topic string) (D, bool) {
	r.mutex.RLock()
	data, ok := r.topicData[topic]
	r.mutex.RUnlock()
	return data, ok
}

func (r *Resolver[D]) DeleteTopicData(topic string) {
	r.mutex.Lock()
	delete(r.topicData, topic)
	r.mutex.Unlock()
}
