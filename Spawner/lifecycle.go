package Spawner

import (
	"Systemge/Error"
	"Systemge/Node"
)

func (spawner *Spawner) OnStop(node *Node.Node) error {
	spawner.mutex.Lock()
	defer spawner.mutex.Unlock()
	for id := range spawner.spawnedNodes {
		err := spawner.EndNode(node, id)
		if err != nil {
			node.GetLogger().Error(Error.New("Error stopping node "+id, err).Error(), node.GetMailer())
		}
	}
	return nil
}
