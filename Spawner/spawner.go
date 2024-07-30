package Spawner

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Node"
	"github.com/neutralusername/Systemge/Tools"
)

func (spawner *Spawner) spawnNode(id string) error {
	_, ok := spawner.spawnedNodes[id]
	if ok {
		return Error.New("Node "+id+" already exists", nil)
	}
	nodeConfig := *spawner.spawnerConfig.SpawnedNodeConfig
	nodeConfig.Name = nodeConfig.Name + "-" + id
	newNode := Node.New(&nodeConfig, spawner.newApplicationFunc(id))
	spawner.spawnedNodes[id] = newNode
	if spawner.spawnerConfig.PropagateSpawnedNodeChanges {
		spawner.addNodeChannel <- newNode
	}
	return nil
}

func (spawner *Spawner) despawnNode(id string) error {
	spawnedNode := spawner.spawnedNodes[id]
	if spawnedNode == nil {
		return Error.New("Node "+id+" does not exist", nil)
	}
	if spawnedNode.IsStarted() {
		err := spawnedNode.Stop()
		if err != nil {
			if errorLogger := spawnedNode.GetErrorLogger(); errorLogger != nil {
				errorLogger.Log(Error.New("Failed stopping node "+id, err).Error())
			}
			if mailer := spawner.node.GetMailer(); mailer != nil {
				err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed stopping node "+id, err).Error()))
				if err != nil {
					if errorLogger := spawner.node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed sending mail", err).Error())
					}
				}
			}
		}
	}
	delete(spawner.spawnedNodes, id)
	if spawner.spawnerConfig.PropagateSpawnedNodeChanges {
		spawner.removeNodeChannel <- spawnedNode
	}
	return nil
}

func (spawner *Spawner) stopNode(id string) error {
	spawnedNode := spawner.spawnedNodes[id]
	if spawnedNode == nil {
		return Error.New("Node "+id+" does not exist", nil)
	}
	err := spawnedNode.Stop()
	if err != nil {
		return Error.New("Failed stopping node "+id, err)
	}
	return nil
}

func (spawner *Spawner) startNode(id string) error {
	spawnedNode := spawner.spawnedNodes[id]
	if spawnedNode == nil {
		return Error.New("Node "+id+" does not exist", nil)
	}
	err := spawnedNode.Start()
	if err != nil {
		return Error.New("Failed starting node", err)
	}
	return nil
}
