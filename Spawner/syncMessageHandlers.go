package Spawner

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Node"
)

func (spawner *Spawner) GetSyncMessageHandlers() map[string]Node.SyncMessageHandler {
	return map[string]Node.SyncMessageHandler{
		"endNodeSync": func(node *Node.Node, message *Message.Message) (string, error) {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.EndNode(node, id)
			spawner.mutex.Unlock()
			if err != nil {
				return "", Error.New("Error ending node "+id, err)
			}
			return id, nil
		},
		"startNodeSync": func(node *Node.Node, message *Message.Message) (string, error) {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.StartNode(node, id)
			spawner.mutex.Unlock()
			if err != nil {
				return "", Error.New("Error starting node "+id, err)
			}
			return id, nil
		},
	}
}
