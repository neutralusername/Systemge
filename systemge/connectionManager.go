package systemge

import "github.com/neutralusername/systemge/tools"

// todo think of a (optional) way to integrate this into server for efficient and user friendly typed messaging

type ConnectionManager[T any] struct {
	*tools.ObjectManager[Connection[T]]
}

func (connectionManager *ConnectionManager[T]) Write(data T, timeoutNs int64, ids ...string) {
	MultiWrite(data, timeoutNs, connectionManager.GetBulk(ids...))
}
