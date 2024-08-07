package Spawner

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Node"
)

func (spawner *Spawner) spawnNode(spawnedNodeConfig *Config.NewNode) error {
	newNode := Node.New(spawnedNodeConfig, spawner.newApplicationFunc())
	spawner.spawnedNodes[newNode.GetName()] = newNode
	if spawner.config.PropagateSpawnedNodeChanges {
		spawner.addNodeChannel <- newNode
	}
	return nil
}

func (spawner *Spawner) despawnNode(nodeName string) error {
	spawnedNode := spawner.spawnedNodes[nodeName]
	if spawnedNode == nil {
		return Error.New("Node "+nodeName+" does not exist", nil)
	}
	spawnedNode.Stop()
	delete(spawner.spawnedNodes, nodeName)
	if spawner.config.PropagateSpawnedNodeChanges {
		spawner.removeNodeChannel <- spawnedNode
	}
	return nil
}

func (spawner *Spawner) stopNode(nodeName string) error {
	spawnedNode := spawner.spawnedNodes[nodeName]
	if spawnedNode == nil {
		return Error.New("Node "+nodeName+" does not exist", nil)
	}
	err := spawnedNode.Stop()
	if err != nil {
		return Error.New("Failed stopping node "+nodeName, err)
	}
	return nil
}

func (spawner *Spawner) startNode(nodeName string) error {
	spawnedNode := spawner.spawnedNodes[nodeName]
	if spawnedNode == nil {
		return Error.New("Node "+nodeName+" does not exist", nil)
	}
	err := spawnedNode.Start()
	if err != nil {
		return Error.New("Failed starting node", err)
	}
	return nil
}
