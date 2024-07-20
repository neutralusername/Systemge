package Dashboard

import "Systemge/Node"

type App struct {
	httpPort uint16
	nodes    map[string]*Node.Node
	node     *Node.Node
}

func new(httpPort uint16, nodes ...*Node.Node) *App {
	server := &App{
		httpPort: httpPort,
		nodes:    make(map[string]*Node.Node),
	}
	for _, node := range nodes {
		server.nodes[node.GetName()] = node
	}
	return server
}
