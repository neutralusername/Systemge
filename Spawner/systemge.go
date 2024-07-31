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
const SPAWN_AND_START_NODE_SYNC = "spawnAndStartNodeSync"
const SPAWN_AND_START_NODE_ASYNC = "spawnAndStartNodeAsync"
const STOP_AND_DESPAWN_NODE_SYNC = "stopAndDespawnNodeSync"
const STOP_AND_DESPAWN_NODE_ASYNC = "stopAndDespawnNodeAsync"

func (spawner *Spawner) GetSyncMessageHandlers() map[string]Node.SyncMessageHandler {
	return map[string]Node.SyncMessageHandler{
		SPAWN_NODE_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			spawner.mutex.Lock()
			err := spawner.spawnNode(Config.UnmarshalNewNode(message.GetPayload()))
			spawner.mutex.Unlock()
			if err != nil {
				return "", Error.New("Failed spawning node \""+message.GetPayload()+"\"", err)
			}
			return "", nil
		},
		DESPAWN_NODE_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			nodeName := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.despawnNode(nodeName)
			spawner.mutex.Unlock()
			if err != nil {
				return "", Error.New("Failed despawning node "+nodeName, err)
			}
			return nodeName, nil
		},
		STOP_NODE_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			nodeName := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.stopNode(nodeName)
			spawner.mutex.Unlock()
			if err != nil {
				return "", Error.New("Failed ending node "+nodeName, err)
			}
			return nodeName, nil
		},
		START_NODE_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			nodeName := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.startNode(nodeName)
			spawner.mutex.Unlock()
			if err != nil {
				return "", Error.New("Failed starting node "+nodeName, err)
			}
			return nodeName, nil
		},
		SPAWN_AND_START_NODE_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			newNodeConfig := Config.UnmarshalNewNode(message.GetPayload())
			spawner.mutex.Lock()
			err := spawner.spawnNode(newNodeConfig)
			if err != nil {
				spawner.mutex.Unlock()
				return "", Error.New("Failed spawning node \""+message.GetPayload()+"\"", err)
			}
			err = spawner.startNode(newNodeConfig.NodeConfig.Name)
			spawner.mutex.Unlock()
			if err != nil {
				return "", Error.New("Failed starting node "+newNodeConfig.NodeConfig.Name, err)
			}
			return "", nil
		},
		STOP_AND_DESPAWN_NODE_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			nodeName := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.despawnNode(nodeName)
			if err != nil {
				spawner.mutex.Unlock()
				return "", Error.New("Failed despawning node "+nodeName, err)
			}
			err = spawner.stopNode(nodeName)
			spawner.mutex.Unlock()
			if err != nil {
				return "", Error.New("Failed ending node "+nodeName, err)
			}
			return nodeName, nil
		},
	}
}

func (spawner *Spawner) GetAsyncMessageHandlers() map[string]Node.AsyncMessageHandler {
	return map[string]Node.AsyncMessageHandler{
		SPAWN_NODE_ASYNC: func(node *Node.Node, message *Message.Message) error {
			spawner.mutex.Lock()
			err := spawner.spawnNode(Config.UnmarshalNewNode(message.GetPayload()))
			spawner.mutex.Unlock()
			if err != nil {
				return Error.New("Failed spawning node \""+message.GetPayload()+"\"", err)
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
		SPAWN_AND_START_NODE_ASYNC: func(node *Node.Node, message *Message.Message) error {
			newNodeConfig := Config.UnmarshalNewNode(message.GetPayload())
			spawner.mutex.Lock()
			err := spawner.spawnNode(newNodeConfig)
			if err != nil {
				spawner.mutex.Unlock()
				return Error.New("Failed spawning node \""+message.GetPayload()+"\"", err)
			}
			err = spawner.startNode(newNodeConfig.NodeConfig.Name)
			spawner.mutex.Unlock()
			if err != nil {
				return Error.New("Failed starting node "+newNodeConfig.NodeConfig.Name, err)
			}
			return nil
		},
		STOP_AND_DESPAWN_NODE_ASYNC: func(node *Node.Node, message *Message.Message) error {
			id := message.GetPayload()
			spawner.mutex.Lock()
			err := spawner.despawnNode(id)
			if err != nil {
				spawner.mutex.Unlock()
				return Error.New("Failed despawning node "+id, err)
			}
			err = spawner.stopNode(id)
			spawner.mutex.Unlock()
			if err != nil {
				return Error.New("Failed ending node "+id, err)
			}
			return nil
		},
	}
}
