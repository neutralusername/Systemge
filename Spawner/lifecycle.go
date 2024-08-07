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
	for nodeName := range spawner.nodes {
		err := spawner.despawnNode(nodeName)
		if err != nil {
			if errorLogger := node.GetErrorLogger(); errorLogger != nil {
				errorLogger.Log(Error.New("Failed despawning spawned node with id \""+nodeName+"\"", err).Error())
			}
			if mailer := node.GetMailer(); mailer != nil {
				err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed despawning spawned node with id \""+nodeName+"\"", err).Error()))
				if err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed sending mail", err).Error())
					}
				}
			}
		}
	}
	return nil
}
