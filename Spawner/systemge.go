package Spawner

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Node"
)

func (spawner *Spawner) OnStart(node *Node.Node) error {
	return nil
}

func (spawner *Spawner) OnStop(node *Node.Node) error {
	spawner.mutex.Lock()
	defer spawner.mutex.Unlock()
	for id := range spawner.spawnedNodes {
		err := spawner.EndNode(node, id)
		if err != nil {
			node.GetLogger().Error(Error.New("Error stopping node "+id, err).Error())
		}
	}
	return nil
}

func (spawner *Spawner) GetSystemgeComponentConfig() Config.Systemge {
	return Config.Systemge{
		HandleMessagesSequentially: false,
	}
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
