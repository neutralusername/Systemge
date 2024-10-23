package systemge

import "github.com/neutralusername/systemge/tools"

type ConnectionManager[D any] struct {
	*tools.ObjectManager[Connection[D]]
}

func (connectionManager *ConnectionManager[D]) Write(data D, timeoutNs int64, connections ...Connection[D]) {
	MultiWrite(data, timeoutNs, connections...)
}

func (connectionManager *ConnectionManager[D]) WriteId(data D, timeoutNs int64, ids ...string) {
	MultiWrite(data, timeoutNs, connectionManager.GetBulk()...)
}
