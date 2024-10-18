package serviceResolver

import (
	"errors"
	"sync"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/serviceSingleRequest"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type Resolver[D any] struct {
	mutex         sync.RWMutex
	topicData     map[string]D
	singleRequest *serviceSingleRequest.SingleRequestServer[D]
	// commands
}

func New[D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterConfig *configs.AccepterServer,
	readerConfig *configs.ReaderServerSync,
	routineConfig *configs.Routine,
	topicData map[string]D,
	handleRequestsConcurrently bool,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	deserializeTopic func(D, systemge.Connection[D]) (string, error), // responsible for retrieving the topic
) (*Resolver[D], error) {

	resolver := &Resolver[D]{
		topicData: topicData,
	}
	singleRequestServer, err := serviceSingleRequest.NewSync(
		listener,
		accepterConfig,
		readerConfig,
		routineConfig,
		handleRequestsConcurrently,
		acceptHandler,
		func(incomingData D, connection systemge.Connection[D]) (D, error) {
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

func (r *Resolver[D]) GetSingleRequest() *serviceSingleRequest.SingleRequestServer[D] {
	return r.singleRequest
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
