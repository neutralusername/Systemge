package Spawner

import (
	"Systemge/Config"
	"Systemge/Node"
	"sync"
)

type Spawner struct {
	appConfig          Config.Application
	spawnerConfig      Config.Spawner
	spawnedNodes       map[string]*Node.Node
	newApplicationFunc func(string) Node.Application
	mutex              sync.Mutex
}

func New(appConfig Config.Application, spawnerConfig Config.Spawner, newApplicationFunc func(string) Node.Application) *Spawner {
	spawner := &Spawner{
		appConfig:          appConfig,
		spawnerConfig:      spawnerConfig,
		spawnedNodes:       make(map[string]*Node.Node),
		newApplicationFunc: newApplicationFunc,
	}
	return spawner
}

func (spawner *Spawner) OnStart(node *Node.Node) error {
	return nil
}

func (spawner *Spawner) OnStop(node *Node.Node) error {
	return nil
}

func (spawner *Spawner) GetApplicationConfig() Config.Application {
	return Config.Application{
		ResolverResolution:         spawner.appConfig.ResolverResolution,
		HandleMessagesSequentially: false,
	}
}
