package Spawner

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Node"
	"github.com/neutralusername/Systemge/Tools"
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
	if spawnedNode.GetStatus() == Node.STATUS_STARTED {
		err := spawnedNode.Stop()
		if err != nil {
			if errorLogger := spawnedNode.GetErrorLogger(); errorLogger != nil {
				errorLogger.Log(Error.New("Failed stopping node "+nodeName, err).Error())
			}
			if mailer := spawner.node.GetMailer(); mailer != nil {
				err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed stopping node "+nodeName, err).Error()))
				if err != nil {
					if errorLogger := spawner.node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed sending mail", err).Error())
					}
				}
			}
		}
	}
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
