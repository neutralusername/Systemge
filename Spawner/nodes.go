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
		err = node.RemoveSyncTopicRemotely(spawner.spawnerConfig.BrokerConfigResolution.GetAddress(), spawner.spawnerConfig.BrokerConfigResolution.GetServerNameIndication(), spawner.spawnerConfig.BrokerConfigResolution.GetTlsCertificate(), id)
		if err != nil {
			node.GetLogger().Log(Error.New("Error removing sync topic \""+id+"\"", err).Error())
		}
	} else {
		err = node.RemoveAsyncTopicRemotely(spawner.spawnerConfig.BrokerConfigResolution.GetAddress(), spawner.spawnerConfig.BrokerConfigResolution.GetServerNameIndication(), spawner.spawnerConfig.BrokerConfigResolution.GetTlsCertificate(), id)
		if err != nil {
			node.GetLogger().Log(Error.New("Error removing async topic \""+id+"\"", err).Error())
		}
	}
	err = node.RemoveResolverTopicRemotely(spawner.spawnerConfig.ResolverConfigResolution.GetAddress(), spawner.spawnerConfig.ResolverConfigResolution.GetServerNameIndication(), spawner.spawnerConfig.ResolverConfigResolution.GetTlsCertificate(), id)
	if err != nil {
		node.GetLogger().Log(Error.New("Error unregistering topic \""+id+"\"", err).Error())
	}
	return nil
}

func (spawner *Spawner) StartNode(node *Node.Node, id string) error {
	if _, ok := spawner.spawnedNodes[id]; ok {
		return Error.New("Node "+id+" already exists", nil)
	}
	newNode := Node.New(Config.Node{
		Name:               id,
		LoggerPath:         spawner.spawnerConfig.SpawnedNodeLoggerPath,
		ResolverResolution: spawner.spawnerConfig.ResolverResolution,
	}, spawner.newApplicationFunc(id))
	if spawner.spawnerConfig.IsSpawnedNodeTopicSync {
		err := node.AddSyncTopicRemotely(spawner.spawnerConfig.BrokerConfigResolution.GetAddress(), spawner.spawnerConfig.BrokerConfigResolution.GetServerNameIndication(), spawner.spawnerConfig.BrokerConfigResolution.GetTlsCertificate(), id)
		if err != nil {
			return Error.New("Error adding sync topic \""+id+"\"", err)
		}
	} else {
		err := node.AddAsyncTopicRemotely(spawner.spawnerConfig.BrokerConfigResolution.GetAddress(), spawner.spawnerConfig.BrokerConfigResolution.GetServerNameIndication(), spawner.spawnerConfig.BrokerConfigResolution.GetTlsCertificate(), id)
		if err != nil {
			return Error.New("Error adding async topic \""+id+"\"", err)
		}
	}
	err := node.AddResolverTopicRemotely(spawner.spawnerConfig.ResolverConfigResolution.GetAddress(), spawner.spawnerConfig.ResolverConfigResolution.GetServerNameIndication(), spawner.spawnerConfig.ResolverConfigResolution.GetTlsCertificate(), spawner.spawnerConfig.BrokerSubscriptionResolution, id)
	if err != nil {
		if spawner.spawnerConfig.IsSpawnedNodeTopicSync {
			removeErr := node.RemoveSyncTopicRemotely(spawner.spawnerConfig.BrokerConfigResolution.GetAddress(), spawner.spawnerConfig.BrokerConfigResolution.GetServerNameIndication(), spawner.spawnerConfig.BrokerConfigResolution.GetTlsCertificate(), id)
			if removeErr != nil {
				node.GetLogger().Log(Error.New("Error removing sync topic \""+id+"\"", removeErr).Error())
			}
		} else {
			removeErr := node.RemoveAsyncTopicRemotely(spawner.spawnerConfig.BrokerConfigResolution.GetAddress(), spawner.spawnerConfig.BrokerConfigResolution.GetServerNameIndication(), spawner.spawnerConfig.BrokerConfigResolution.GetTlsCertificate(), id)
			if removeErr != nil {
				node.GetLogger().Log(Error.New("Error removing async topic \""+id+"\"", removeErr).Error())
			}
		}
		return Error.New("Error registering topic", err)
	}
	err = newNode.Start()
	if err != nil {
		if spawner.spawnerConfig.IsSpawnedNodeTopicSync {
			removeErr := node.RemoveSyncTopicRemotely(spawner.spawnerConfig.BrokerConfigResolution.GetAddress(), spawner.spawnerConfig.BrokerConfigResolution.GetServerNameIndication(), spawner.spawnerConfig.BrokerConfigResolution.GetTlsCertificate(), id)
			if removeErr != nil {
				node.GetLogger().Log(Error.New("Error removing sync topic \""+id+"\"", removeErr).Error())
			}
		} else {
			removeErr := node.RemoveAsyncTopicRemotely(spawner.spawnerConfig.BrokerConfigResolution.GetAddress(), spawner.spawnerConfig.BrokerConfigResolution.GetServerNameIndication(), spawner.spawnerConfig.BrokerConfigResolution.GetTlsCertificate(), id)
			if removeErr != nil {
				node.GetLogger().Log(Error.New("Error removing async topic \""+id+"\"", removeErr).Error())
			}
		}
		removeErr := node.RemoveResolverTopicRemotely(spawner.spawnerConfig.ResolverConfigResolution.GetAddress(), spawner.spawnerConfig.ResolverConfigResolution.GetServerNameIndication(), spawner.spawnerConfig.ResolverConfigResolution.GetTlsCertificate(), id)
		if removeErr != nil {
			node.GetLogger().Log(Error.New("Error unregistering topic \""+id+"\"", removeErr).Error())
		}
		return Error.New("Error starting node", err)
	}
	spawner.spawnedNodes[id] = newNode
	return nil
}
