package servicePublishSubscribe

import (
	"sync"

	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type Client[D any] struct {
	topicResolverIntervals map[string]int64                          // topic -> resolveInterval
	topics                 map[string]map[*connection[D]]int64       // topic -> connections -> subscribe topic resolveInterval
	connections            map[systemge.Connection[D]]*connection[D] // connection -> connection

	resolveFunc func(string) []systemge.Connection[D]

	mutex sync.RWMutex
}

type connection[D any] struct {
	connection systemge.Connection[D]
	topics     map[string]struct{}
	timeout    *tools.Timeout
}

func NewClient[D any](
	topicResolverIntervals map[string]int64,
	resolveFunc func(string) []systemge.Connection[D],
	resolveOnConnectionLoss bool,
	topicResolveInterval int64,
) {
	client := &Client[D]{
		topicResolverIntervals: topicResolverIntervals,
		topics:                 make(map[string]map[*connection[D]]int64),
		connections:            make(map[systemge.Connection[D]]*connection[D]),
		resolveFunc:            resolveFunc,
	}
	for topic, _ := range topicResolverIntervals {
		client.topics[topic] = make(map[*connection[D]]int64)
	}

}

func (client *Client[D]) Start() error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	for topic, connections := range client.topics {
		/* for _, connection_ := range client.resolveFunc(topic) {
			connection := &connection[D]{
				connection: connection_,
				timeout: tools.NewTimeout(
					client.topicResolverIntervals[topic],
					func() {

					},
					false),
			}
			connections[connection] = client.topicResolverIntervals[topic]
			client.connections[connection_] = connection
		} */
	}

	return nil
}
