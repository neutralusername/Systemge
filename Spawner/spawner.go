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
		Name:          id,
		InfoLogger:    Tools.NewLogger("[Info \""+id+"\"]", spawner.spawnerConfig.LoggerQueue),
		WarningLogger: Tools.NewLogger("[Warning \""+id+"\"] ", spawner.spawnerConfig.LoggerQueue),
		ErrorLogger:   Tools.NewLogger("[Error \""+id+"\"] ", spawner.spawnerConfig.LoggerQueue),
		DebugLogger:   Tools.NewLogger("[Debug \""+id+"\"] ", spawner.spawnerConfig.LoggerQueue),
		Mailer:        spawner.spawnerConfig.Mailer,
	}
	if spawner.spawnerConfig.LogInternals {
		node.InternalInfoLogger = Tools.NewLogger("[InternalInfo \""+id+"\"]", spawner.spawnerConfig.LoggerQueue)
		node.InternalWarningLogger = Tools.NewLogger("[InternalWarning \""+id+"\"] ", spawner.spawnerConfig.LoggerQueue)
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
	}
	if spawner.spawnerConfig.IsSpawnedNodeTopicSync {
		removeErr := spawner.node.RemoveSyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
		if removeErr != nil {
			if errorLogger := spawner.node.GetErrorLogger(); errorLogger != nil {
				errorLogger.Log(Error.New("Failed removing sync topic \""+id+"\"", removeErr).Error())
				if mailer := spawner.node.GetMailer(); mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed removing sync topic \""+id+"\"", removeErr).Error()))
					if err != nil {
						if errorLogger := spawner.node.GetErrorLogger(); errorLogger != nil {
							errorLogger.Log(Error.New("Failed sending mail", err).Error())
						}
					}
				}
			}
		}
	} else {
		removeErr := spawner.node.RemoveAsyncTopicRemotely(spawner.spawnerConfig.BrokerConfigEndpoint, id)
		if removeErr != nil {
			if errorLogger := spawner.node.GetErrorLogger(); errorLogger != nil {
				errorLogger.Log(Error.New("Failed removing async topic \""+id+"\"", removeErr).Error())
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
