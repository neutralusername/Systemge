package servicePublishSubscribe

import (
	"github.com/neutralusername/systemge/systemge"
)

type Client[D any] struct {
	connection systemge.Connection[D]
}

func NewPublishSubscribeClient[D any](
	connection systemge.Connection[D],
) (*Client[D], error) {

	client := &Client[D]{
		connection: connection,
	}

	return client, nil
}
