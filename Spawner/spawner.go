package Spawner

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Node"
	"github.com/neutralusername/Systemge/Tools"
)

func (spawner *Spawner) spawnNode(id string) error {
	_, ok := spawner.spawnedNodes[id]
	if ok {
		return Error.New("Node "+id+" already exists", nil)
	}
	node := &Config.Node{
		Name:                      id,
		InfoLoggerPath:            spawner.spawnerConfig.InfoLoggerPath,
		InternalInfoLoggerPath:    spawner.spawnerConfig.InternalLoggerPath,
		WarningLoggerPath:         spawner.spawnerConfig.WarningLoggerPath,
		InternalWarningLoggerPath: spawner.spawnerConfig.InternalWarningLoggerPath,
		ErrorLoggerPath:           spawner.spawnerConfig.ErrorLoggerPath,
		DebugLoggerPath:           spawner.spawnerConfig.DebugLoggerPath,
		Mailer:                    spawner.spawnerConfig.Mailer,
	}
	newNode := Node.New(node, spawner.newApplicationFunc(id))
	if spawner.spawnerConfig.IsSpawnedNodeTopicSync {
		err := spawner.node.AddSyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
		if err != nil {
			return Error.New("Failed adding sync topic \""+id+"\"", err)
		}
	} else {
		err := spawner.node.AddAsyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
		if err != nil {
			return Error.New("Failed adding async topic \""+id+"\"", err)
		}
	}
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
	if spawner.spawnerConfig.IsSpawnedNodeTopicSync {
		removeErr := spawner.node.RemoveSyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
		if removeErr != nil {
			if errorLogger := spawner.node.GetErrorLogger(); errorLogger != nil {
				errorLogger.Log(Error.New("Failed removing sync topic \""+id+"\"", removeErr).Error())
			}
			if mailer := spawner.node.GetMailer(); mailer != nil {
				err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed removing sync topic \""+id+"\"", removeErr).Error()))
				if err != nil {
					if errorLogger := spawner.node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed sending mail", err).Error())
					}
				}
			}
		}
	} else {
		removeErr := spawner.node.RemoveAsyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
		if removeErr != nil {
			if errorLogger := spawner.node.GetErrorLogger(); errorLogger != nil {
				errorLogger.Log(Error.New("Failed removing async topic \""+id+"\"", removeErr).Error())
			}
			if mailer := spawner.node.GetMailer(); mailer != nil {
				err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed removing async topic \""+id+"\"", removeErr).Error()))
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
