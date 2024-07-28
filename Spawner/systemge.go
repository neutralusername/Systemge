package Spawner

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Node"
)

const START_NODE_SYNC = "startNodeSync"
const START_NODE_ASYNC = "startNodeAsync"
const STOP_NODE_SYNC = "stopNodeSync"
const STOP_NODE_ASYNC = "stopNodeAsync"
const SPAWN_NODE_SYNC = "spawnNodeSync"
const SPAWN_NODE_ASYNC = "spawnNodeAsync"
const DESPAWN_NODE_SYNC = "despawnNodeSync"
const DESPAWN_NODE_ASYNC = "despawnNodeAsync"

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
				return "", Error.New("Failed spawning node "+id, err)
			}
			return id, nil
		},
		DESPAWN_NODE_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.despawnNode(id)
			spawner.mutex.Unlock()
			if err != nil {
				return "", Error.New("Failed despawning node "+id, err)
			}
			return id, nil
		},
		STOP_NODE_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.stopNode(id)
			spawner.mutex.Unlock()
			if err != nil {
				return "", Error.New("Failed ending node "+id, err)
			}
			return id, nil
		},
		START_NODE_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.startNode(id)
			spawner.mutex.Unlock()
			if err != nil {
				return "", Error.New("Failed starting node "+id, err)
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
				return Error.New("Failed spawning node "+id, err)
			}
			return nil
		},
		DESPAWN_NODE_ASYNC: func(node *Node.Node, message *Message.Message) error {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.despawnNode(id)
			spawner.mutex.Unlock()
			if err != nil {
				return Error.New("Failed despawning node "+id, err)
			}
			return nil
		},
		STOP_NODE_ASYNC: func(node *Node.Node, message *Message.Message) error {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.stopNode(id)
			spawner.mutex.Unlock()
			if err != nil {
				return Error.New("Failed ending node "+id, err)
			}
			return nil
		},
		START_NODE_ASYNC: func(node *Node.Node, message *Message.Message) error {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.startNode(id)
			spawner.mutex.Unlock()
			if err != nil {
				return Error.New("Failed starting node "+id, err)
			}
			return nil
		},
	}
}
