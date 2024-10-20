package servicePublishSubscribe

import (
	"sync"

	"github.com/neutralusername/systemge/systemge"
)

type Client[D any] struct {
	topicResolverIntervals map[string]int64                            // topic -> resolveInterval
	topics                 map[string]map[systemge.Connection[D]]int64 // topic -> connections -> subscribe topic resolveInterval

	resolveFunc func(string) []systemge.Connection[D]

	mutex sync.RWMutex
}

func NewClient[D any](
	topicResolverIntervals map[string]int64,
	resolveFunc func(string) []systemge.Connection[D],
	resolveOnConnectionLoss bool,
	topicResolveInterval int64,
) {
	client := &Client[D]{
		topicResolverIntervals: topicResolverIntervals,
		topics:                 make(map[string]map[systemge.Connection[D]]int64),
		resolveFunc:            resolveFunc,
	}
	for topic, _ := range topicResolverIntervals {
		client.topics[topic] = make(map[systemge.Connection[D]]int64)
	}

}

func (client *Client[D]) Start() error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	for topic, connections := range client.topics {
		resolveConnections := client.resolveFunc(topic)
		for _, connection := range resolveConnections {
			connections[connection] = client.topicResolverIntervals[topic]
		}
	}

	return nil
}
