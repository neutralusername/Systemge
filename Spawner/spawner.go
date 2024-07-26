package Spawner

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Node"
	"Systemge/Tools"
)

func (spawner *Spawner) spawnNode(id string) error {
	_, ok := spawner.spawnedNodes[id]
	if ok {
		return Error.New("Node "+id+" already exists", nil)
	}
	newNode := Node.New(&Config.Node{
		Name:          id,
		ErrorLogger:   spawner.spawnerConfig.ErrorLogger,
		WarningLogger: spawner.spawnerConfig.WarningLogger,
		InfoLogger:    spawner.spawnerConfig.InfoLogger,
		DebugLogger:   spawner.spawnerConfig.DebugLogger,
		Mailer:        spawner.spawnerConfig.Mailer,
	}, spawner.newApplicationFunc(id))
	if spawner.spawnerConfig.IsSpawnedNodeTopicSync {
		err := spawner.node.AddSyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
		if err != nil {
			return Error.New("Error adding sync topic \""+id+"\"", err)
		}
	} else {
		err := spawner.node.AddAsyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
		if err != nil {
			return Error.New("Error adding async topic \""+id+"\"", err)
		}
	}
	spawner.spawnedNodes[id] = newNode
	spawner.addNodeChannel <- newNode
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
				errorLogger.Log(Error.New("Error stopping node "+id, err).Error())
				if mailer := spawner.node.GetMailer(); mailer != nil {
					mailer.Send(Tools.NewMail(nil, "error", Error.New("Error stopping node "+id, err).Error()))
				}
			}
		}
	}
	if spawner.spawnerConfig.IsSpawnedNodeTopicSync {
		removeErr := spawner.node.RemoveSyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
		if removeErr != nil {
			if errorLogger := spawner.node.GetErrorLogger(); errorLogger != nil {
				return Error.New("Error removing sync topic \""+id+"\"", removeErr)
			}
		}
	} else {
		removeErr := spawner.node.RemoveAsyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
		if removeErr != nil {
			if errorLogger := spawner.node.GetErrorLogger(); errorLogger != nil {
				return Error.New("Error removing async topic \""+id+"\"", removeErr)
			}
		}
	}
	delete(spawner.spawnedNodes, id)
	spawner.removeNodeChannel <- spawnedNode
	return nil
}

func (spawner *Spawner) stopNode(id string) error {
	spawnedNode := spawner.spawnedNodes[id]
	if spawnedNode == nil {
		return Error.New("Node "+id+" does not exist", nil)
	}
	err := spawnedNode.Stop()
	if err != nil {
		return Error.New("Error stopping node "+id, err)
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
		return Error.New("Error starting node", err)
	}
	return nil
}
