package servicePublishSubscribe

import "github.com/neutralusername/systemge/systemge"

type Client[D any] struct {
	topics map[string]map[systemge.Connection[D]]int64 // topic -> connections -> subscribe topic resolveInterval
}

func NewClient[D any](
	topics map[string]int64,
	resolveFunc func(string) []systemge.Connection[D],
	resolveOnConnectionLoss bool,
	topicResolveInterval int64,
) {
	client := &Client[D]{
		topics: make(map[string]map[systemge.Connection[D]]int64),
	}
	for topic, resolveInterval := range topics {
		client.topics[topic] = make(map[systemge.Connection[D]]int64)
	}

}

func (client *Client[D]) Start() error {
	for topic, connections := range client.topics {

	}
}
