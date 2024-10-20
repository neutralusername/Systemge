package servicePublishSubscribe

import "github.com/neutralusername/systemge/systemge"

type Client[D any] struct {
	topicResolverIntervals map[string]int64                            // topic -> resolveInterval
	topics                 map[string]map[systemge.Connection[D]]int64 // topic -> connections -> subscribe topic resolveInterval
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
	}
	for topic, _ := range topicResolverIntervals {
		client.topics[topic] = make(map[systemge.Connection[D]]int64)
	}

}

func (client *Client[D]) Start() error {
	for topic, connections := range client.topics {

	}
}
