package Dashboard

import (
	"Systemge/Config"
	"Systemge/Node"
)

type App struct {
	nodes   map[string]*Node.Node
	node    *Node.Node
	config  *Config.Dashboard
	started bool
}

func New(config *Config.Dashboard, nodes ...*Node.Node) *App {
	server := &App{
		nodes:  make(map[string]*Node.Node),
		config: config,
	}
	for _, node := range nodes {
		server.nodes[node.GetName()] = node
	}
	return server
}
