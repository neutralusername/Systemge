package Spawner

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Node"
)

type Spawner struct {
	config             *Config.Spawner
	nodes              map[string]*Node.Node
	newApplicationFunc func() Node.Application
	mutex              sync.Mutex
	node               *Node.Node
	nodeChangeChannel  chan *SpawnerNodeChange
}

type SpawnerNodeChange struct {
	Node  *Node.Node
	Added bool
}

func New(config *Config.Spawner, newApplicationFunc func() Node.Application) *Node.Node {
	app := &Spawner{
		config:             config,
		nodes:              make(map[string]*Node.Node),
		newApplicationFunc: newApplicationFunc,
		nodeChangeChannel:  make(chan *SpawnerNodeChange),
	}
	node := Node.New(&Config.NewNode{
		NodeConfig:     config.NodeConfig,
		SystemgeConfig: config.SystemgeConfig,
	}, app)
	return node
}

func (spawner *Spawner) GetNextNodeChange() *SpawnerNodeChange {
	return <-spawner.nodeChangeChannel
}

func ImplementsSpawner(application Node.Application) bool {
	_, ok := application.(*Spawner)
	return ok
}

func (spawner *Spawner) GetSpawnedNodeCount() int {
	spawner.mutex.Lock()
	defer spawner.mutex.Unlock()
	return len(spawner.nodes)
}
