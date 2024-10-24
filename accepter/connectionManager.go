package accepter

import (
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type ConnectionManager[T any] struct {
	*tools.ObjectManager[systemge.Connection[T]]
}

func NewConnectionManager[T any](
	objectManager *tools.ObjectManager[systemge.Connection[T]],
) *ConnectionManager[T] {
	return &ConnectionManager[T]{
		ObjectManager: objectManager,
	}
}

func (connectionManager *ConnectionManager[T]) Write(data T, timeoutNs int64, ids ...string) {
	systemge.MultiWrite(data, timeoutNs, connectionManager.GetBulk(ids...))
}
