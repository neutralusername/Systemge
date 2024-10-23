package systemge

import "github.com/neutralusername/systemge/tools"

// todo think of a way to integrate this with server for effective and user friendly typed messaging

type ConnectionManager[D any] struct {
	*tools.ObjectManager[Connection[D]]
}

func (connectionManager *ConnectionManager[D]) Write(data D, timeoutNs int64, ids ...string) {
	MultiWrite(data, timeoutNs, connectionManager.GetBulk(ids...))
}
