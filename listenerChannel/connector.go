package listenerChannel

import (
	"github.com/neutralusername/systemge/connectionChannel"
	"github.com/neutralusername/systemge/systemge"
)

type connector[D any] struct {
	connChann chan *connectionChannel.ConnectionRequest[D]
}

func NewConnector[D any](
	connChann chan *connectionChannel.ConnectionRequest[D],
) systemge.Connector[D] {
	return &connector[D]{
		connChann: connChann,
	}
}

func (connector *connector[D]) Connect(timeoutNs int64) (systemge.Connection[D], error) {
	return Connect(connector.connChann, timeoutNs)
}
