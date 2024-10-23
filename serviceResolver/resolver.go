package serviceResolver

import (
	"errors"
	"sync"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/serviceSingleRequest"
	"github.com/neutralusername/systemge/systemge"
)

type Resolver[T any] struct {
	mutex         sync.RWMutex
	topicData     map[string]T
	singleRequest *serviceSingleRequest.SingleRequestServer[T]
	// commands
}

func New[T any](
	listener systemge.Listener[T],
	topicData map[string]T,
	accepterConfig *configs.Accepter,
	readerSyncConfig *configs.ReaderSync,
	routineConfig *configs.Routine,
	acceptHandler systemge.AcceptHandlerWithError[T],
	deserializeTopic func(T, systemge.Connection[T]) (string, error), // responsible for retrieving the topic
) (*Resolver[T], error) {

	if deserializeTopic == nil {
		return nil, errors.New("deserializeTopic is nil")
	}

	resolver := &Resolver[T]{
		topicData: topicData,
	}

	singleRequestServer, err := serviceSingleRequest.NewSync(
		listener,
		accepterConfig,
		readerSyncConfig,
		routineConfig,
		acceptHandler,
		func(incomingData T, connection systemge.Connection[T]) (T, error) {
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

	resolver.singleRequest = singleRequestServer
	return resolver, nil
}

func (r *Resolver[T]) GetSingleRequest() *serviceSingleRequest.SingleRequestServer[T] {
	return r.singleRequest
}

func (r *Resolver[T]) SetTopicData(topic string, data T) {
	r.mutex.Lock()
	r.topicData[topic] = data
	r.mutex.Unlock()
}

func (r *Resolver[T]) GetTopicData(topic string) (T, bool) {
	r.mutex.RLock()
	data, ok := r.topicData[topic]
	r.mutex.RUnlock()
	return data, ok
}

func (r *Resolver[T]) DeleteTopicData(topic string) {
	r.mutex.Lock()
	delete(r.topicData, topic)
	r.mutex.Unlock()
}
