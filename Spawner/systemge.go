package Spawner

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Node"
)

func (spawner *Spawner) GetSystemgeComponentConfig() Config.Systemge {
	return spawner.systemgeConfig
}

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

func (spawner *Spawner) GetAsyncMessageHandlers() map[string]Node.AsyncMessageHandler {
	return map[string]Node.AsyncMessageHandler{
		"endNodeAsync": func(node *Node.Node, message *Message.Message) error {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.EndNode(node, id)
			spawner.mutex.Unlock()
			if err != nil {
				return Error.New("Error ending node "+id, err)
			}
			return nil
		},
		"startNodeAsync": func(node *Node.Node, message *Message.Message) error {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.StartNode(node, id)
			spawner.mutex.Unlock()
			if err != nil {
				return Error.New("Error starting node "+id, err)
			}
			return nil
		},
	}
}
