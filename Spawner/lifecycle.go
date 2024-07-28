package Spawner

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Node"
	"github.com/neutralusername/Systemge/Tools"
)

func (spawner *Spawner) OnStart(node *Node.Node) error {
	spawner.node = node
	return nil
}

func (spawner *Spawner) OnStop(node *Node.Node) error {
	spawner.mutex.Lock()
	defer spawner.mutex.Unlock()
	for id := range spawner.spawnedNodes {
		err := spawner.despawnNode(id)
		if err != nil {
			if errorLogger := node.GetErrorLogger(); errorLogger != nil {
				errorLogger.Log(Error.New("Error despawning spawned node with id \""+id+"\"", err).Error())
				if mailer := node.GetMailer(); mailer != nil {
					mailer.Send(Tools.NewMail(nil, "error", Error.New("Error despawning spawned node with id \""+id+"\"", err).Error()))
				}
			}
		}
	}
	return nil
}
