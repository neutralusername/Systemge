package Spawner

import (
	"Systemge/Config"
	"Systemge/Node"
	"sync"
)

type Spawner struct {
	spawnerConfig      *Config.Spawner
	systemgeConfig     *Config.Systemge
	spawnedNodes       map[string]*Node.Node
	newApplicationFunc func(string) Node.Application
	mutex              sync.Mutex
}

func New(spawnerConfig *Config.Spawner, systemgeConfig *Config.Systemge, newApplicationFunc func(string) Node.Application) *Spawner {
	spawner := &Spawner{
		spawnerConfig:      spawnerConfig,
		systemgeConfig:     systemgeConfig,
		spawnedNodes:       make(map[string]*Node.Node),
		newApplicationFunc: newApplicationFunc,
	}
	return spawner
}

func ImplementsSpawner(node *Node.Node) bool {
	_, ok := node.GetApplication().(*Spawner)
	return ok
}

func (spawner *Spawner) GetSpawnedNodeCount() int {
	spawner.mutex.Lock()
	defer spawner.mutex.Unlock()
	return len(spawner.spawnedNodes)
}
