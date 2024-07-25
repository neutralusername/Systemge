package Spawner

import (
	"Systemge/Error"
	"Systemge/Node"
)

func (spawner *Spawner) OnStart(node *Node.Node) error {
	spawner.node = node
	return nil
}

func (spawner *Spawner) OnStop(node *Node.Node) error {
	spawner.mutex.Lock()
	defer spawner.mutex.Unlock()
	for id := range spawner.spawnedNodes {
		err := spawner.stopNode(id)
		if err != nil {
			if errorLogger := node.GetErrorLogger(); errorLogger != nil {
				errorLogger.Log(Error.New("Error stopping spawned node with id \""+id+"\"", err).Error(), node.GetMailer())
			}
		}
		err = spawner.despawnNode(id)
		if err != nil {
			if errorLogger := node.GetErrorLogger(); errorLogger != nil {
				errorLogger.Log(Error.New("Error despawning spawned node with id \""+id+"\"", err).Error(), node.GetMailer())
			}
		}
	}
	close(spawner.addNodeChannel)
	close(spawner.removeNodeChannel)
	return nil
}
