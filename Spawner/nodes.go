package Spawner

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Node"
)

func (spawner *Spawner) EndNode(node *Node.Node, id string) error {
	spawnedNode := spawner.spawnedNodes[id]
	if spawnedNode == nil {
		return Error.New("Node "+id+" does not exist", nil)
	}
	err := spawnedNode.Stop()
	if err != nil {
		return Error.New("Error stopping node "+id, err)
	}
	delete(spawner.spawnedNodes, id)
	if spawner.spawnerConfig.IsSpawnedNodeTopicSync {
		err = node.RemoveSyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
		if err != nil {
			node.GetLogger().Info(Error.New("Error removing sync topic \""+id+"\"", err).Error())
		}
	} else {
		err = node.RemoveAsyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
		if err != nil {
			node.GetLogger().Info(Error.New("Error removing async topic \""+id+"\"", err).Error())
		}
	}
	return nil
}

func (spawner *Spawner) StartNode(node *Node.Node, id string) error {
	if _, ok := spawner.spawnedNodes[id]; ok {
		return Error.New("Node "+id+" already exists", nil)
	}
	newNode := Node.New(Config.Node{
		Name:                      id,
		Logger:                    spawner.spawnerConfig.SpawnedNodeLogger,
		ResolverEndpoint:          spawner.spawnerConfig.ResolverEndpoint,
		SyncResponseTimeoutMs:     1000,
		TopicResolutionLifetimeMs: 10000,
		BrokerSubscribeDelayMs:    1000,
	}, spawner.newApplicationFunc(id))
	if spawner.spawnerConfig.IsSpawnedNodeTopicSync {
		err := node.AddSyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
		if err != nil {
			return Error.New("Error adding sync topic \""+id+"\"", err)
		}
	} else {
		err := node.AddAsyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
		if err != nil {
			return Error.New("Error adding async topic \""+id+"\"", err)
		}
	}
	err := newNode.Start()
	if err != nil {
		if spawner.spawnerConfig.IsSpawnedNodeTopicSync {
			removeErr := node.RemoveSyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
			if removeErr != nil {
				node.GetLogger().Info(Error.New("Error removing sync topic \""+id+"\"", removeErr).Error())
			}
		} else {
			removeErr := node.RemoveAsyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
			if removeErr != nil {
				node.GetLogger().Info(Error.New("Error removing async topic \""+id+"\"", removeErr).Error())
			}
		}
		return Error.New("Error starting node", err)
	}
	spawner.spawnedNodes[id] = newNode
	return nil
}
