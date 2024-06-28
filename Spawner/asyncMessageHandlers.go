package Spawner

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Node"
)

func (spawner *Spawner) GetAsyncMessageHandlers() map[string]Node.AsyncMessageHandler {
	return map[string]Node.AsyncMessageHandler{
		"endNodeAsync": func(node *Node.Node, message *Message.Message) error {
			id := message.GetPayload()
			err := spawner.EndNode(node, id)
			if err != nil {
				return Error.New("Error ending node "+id, err)
			}
			return nil
		},
		"startNodeAsync": func(node *Node.Node, message *Message.Message) error {
			id := message.GetPayload()
			err := spawner.StartNode(node, id)
			if err != nil {
				return Error.New("Error starting node "+id, err)
			}
			return nil
		},
	}
}
