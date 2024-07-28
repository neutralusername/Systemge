package Dashboard

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Node"
)

type App struct {
	nodes   map[string]*Node.Node
	node    *Node.Node
	config  *Config.Dashboard
	started bool
}

func New(config *Config.Dashboard, nodes ...*Node.Node) *App {
	app := &App{
		nodes:  make(map[string]*Node.Node),
		config: config,
	}

	for _, node := range nodes {
		if app.config.AutoStart {
			err := node.Start()
			if err != nil {
				panic(err)
			}
		}
		app.nodes[node.GetName()] = node
	}
	return app
}
