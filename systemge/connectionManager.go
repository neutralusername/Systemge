package systemge

import "github.com/neutralusername/systemge/tools"

type ConnectionManager[D any] struct {
	*tools.ObjectManager[Connection[D]]
}

func (connectionManager *ConnectionManager[D]) Write(data D, timeoutNs int64, connections ...Connection[D]) {
	if len(connections) == 0 {
		connections = connectionManager.GetBulk()
	}
	MultiWrite(data, timeoutNs, connections...)
}

func (connectionManager *ConnectionManager[D]) WriteId(data D, timeoutNs int64, ids ...string) {
	MultiWrite(data, timeoutNs, connectionManager.GetBulk()...)
}
