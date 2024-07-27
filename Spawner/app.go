package Spawner

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Node"
)

type Spawner struct {
	spawnerConfig      *Config.Spawner
	systemgeConfig     *Config.Systemge
	spawnedNodes       map[string]*Node.Node
	newApplicationFunc func(string) Node.Application
	mutex              sync.Mutex
	node               *Node.Node
	addNodeChannel     chan *Node.Node
	removeNodeChannel  chan *Node.Node
}

func New(spawnerConfig *Config.Spawner, systemgeConfig *Config.Systemge, newApplicationFunc func(string) Node.Application) *Spawner {
	spawner := &Spawner{
		spawnerConfig:      spawnerConfig,
		systemgeConfig:     systemgeConfig,
		spawnedNodes:       make(map[string]*Node.Node),
		newApplicationFunc: newApplicationFunc,
		addNodeChannel:     make(chan *Node.Node),
		removeNodeChannel:  make(chan *Node.Node),
	}
	return spawner
}

func (spawner *Spawner) GetAddNodeChannel() chan *Node.Node {
	return spawner.addNodeChannel
}

func (spawner *Spawner) GetRemoveNodeChannel() chan *Node.Node {
	return spawner.removeNodeChannel
}

func ImplementsSpawner(application Node.Application) bool {
	_, ok := application.(*Spawner)
	return ok
}

func (spawner *Spawner) GetSpawnedNodeCount() int {
	spawner.mutex.Lock()
	defer spawner.mutex.Unlock()
	return len(spawner.spawnedNodes)
}
