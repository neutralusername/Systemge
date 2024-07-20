package Dashboard

import "Systemge/Node"

type App struct {
	httpPort uint16
	nodes    []*Node.Node
	node     *Node.Node
}

func new(httpPort uint16, nodes ...*Node.Node) *App {
	server := &App{
		httpPort: httpPort,
		nodes:    nodes,
	}
	return server
}
