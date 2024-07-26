package Spawner

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Node"
)

func (spawner *Spawner) GetSystemgeComponentConfig() *Config.Systemge {
	return spawner.systemgeConfig
}

func (spawner *Spawner) GetSyncMessageHandlers() map[string]Node.SyncMessageHandler {
	return map[string]Node.SyncMessageHandler{
		SPAWN_NODE_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.spawnNode(id)
			spawner.mutex.Unlock()
			if err != nil {
				return "", Error.New("Error spawning node "+id, err)
			}
			return id, nil
		},
		DESPAWN_NODE_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.despawnNode(id)
			spawner.mutex.Unlock()
			if err != nil {
				return "", Error.New("Error despawning node "+id, err)
			}
			return id, nil
		},
		STOP_NODE_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.stopNode(id)
			spawner.mutex.Unlock()
			if err != nil {
				return "", Error.New("Error ending node "+id, err)
			}
			return id, nil
		},
		START_NODE_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.startNode(id)
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
		SPAWN_NODE_ASYNC: func(node *Node.Node, message *Message.Message) error {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.spawnNode(id)
			spawner.mutex.Unlock()
			if err != nil {
				return Error.New("Error spawning node "+id, err)
			}
			return nil
		},
		DESPAWN_NODE_ASYNC: func(node *Node.Node, message *Message.Message) error {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.despawnNode(id)
			spawner.mutex.Unlock()
			if err != nil {
				return Error.New("Error despawning node "+id, err)
			}
			return nil
		},
		STOP_NODE_ASYNC: func(node *Node.Node, message *Message.Message) error {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.stopNode(id)
			spawner.mutex.Unlock()
			if err != nil {
				return Error.New("Error ending node "+id, err)
			}
			return nil
		},
		START_NODE_ASYNC: func(node *Node.Node, message *Message.Message) error {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.startNode(id)
			spawner.mutex.Unlock()
			if err != nil {
				return Error.New("Error starting node "+id, err)
			}
			return nil
		},
	}
}
