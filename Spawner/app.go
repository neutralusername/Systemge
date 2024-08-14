package Spawner

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Node"
	"github.com/neutralusername/Systemge/Tools"
)

type Spawner struct {
	config             *Config.Spawner
	nodes              map[string]*Node.Node
	spawnedNodesCount  int
	newApplicationFunc func() Node.Application
	mutex              sync.Mutex
	nodeChangeChannel  chan *SpawnerNodeChange
	errorLogger        *Tools.Logger
}

type SpawnerNodeChange struct {
	Node  *Node.Node
	Added bool
}

func New(config *Config.Spawner, newApplicationFunc func() Node.Application) *Spawner {
	app := &Spawner{
		config:             config,
		nodes:              make(map[string]*Node.Node),
		newApplicationFunc: newApplicationFunc,
		nodeChangeChannel:  make(chan *SpawnerNodeChange),
		errorLogger:        Tools.NewLogger("[Error: \"Spawner\"]", config.ErrorLoggerPath),
	}
	return app
}

func (spawner *Spawner) GetNextNodeChange() *SpawnerNodeChange {
	return <-spawner.nodeChangeChannel
}

func ImplementsSpawner(application Node.Application) bool {
	_, ok := application.(*Spawner)
	return ok
}

func (spawner *Spawner) GetSpawnedNodeCount() int {
	return spawner.spawnedNodesCount
}
