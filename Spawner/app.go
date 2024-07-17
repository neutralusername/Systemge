package Spawner

import (
	"Systemge/Config"
	"Systemge/Node"
	"sync"
)

type Spawner struct {
	spawnerConfig      Config.Spawner
	spawnedNodes       map[string]*Node.Node
	newApplicationFunc func(string) Node.Application
	mutex              sync.Mutex
}

func New(spawnerConfig Config.Spawner, newApplicationFunc func(string) Node.Application) *Spawner {
	spawner := &Spawner{
		spawnerConfig:      spawnerConfig,
		spawnedNodes:       make(map[string]*Node.Node),
		newApplicationFunc: newApplicationFunc,
	}
	return spawner
}
