package listenerChannel

import (
	"github.com/neutralusername/systemge/connectionChannel"
	"github.com/neutralusername/systemge/systemge"
)

type connector[T any] struct {
	connChann chan *connectionChannel.ConnectionRequest[T]
}

func NewConnector[T any](
	connChann chan *connectionChannel.ConnectionRequest[T],
) systemge.Connector[T] {
	return &connector[T]{
		connChann: connChann,
	}
}

func (connector *connector[T]) Connect(timeoutNs int64) (systemge.Connection[T], error) {
	return Connect(connector.connChann, timeoutNs)
}
