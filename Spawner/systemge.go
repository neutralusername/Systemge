package Spawner

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Node"
	"github.com/neutralusername/Systemge/Tools"
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
const STOP_ALL_NODES_SYNC = "stopAllNodesSync"
const STOP_ALL_NODES_ASYNC = "stopAllNodesAsync"
const DESPAWN_ALL_NODES_SYNC = "despawnAllNodesSync"
const DESPAWN_ALL_NODES_ASYNC = "despawnAllNodesAsync"

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
				return "", Error.New("Failed stopping node "+nodeName, err)
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

			if err != nil {
				err = spawner.despawnNode(newNodeConfig.NodeConfig.Name)
				if err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed despawning node "+newNodeConfig.NodeConfig.Name, err).Error())
					}
					if mailer := node.GetMailer(); mailer != nil {
						err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed despawning node "+newNodeConfig.NodeConfig.Name, err).Error()))
						if err != nil {
							if errorLogger := node.GetErrorLogger(); errorLogger != nil {
								errorLogger.Log(Error.New("Failed sending mail", err).Error())
							}
						}
					}
				}
				spawner.mutex.Unlock()
				return "", Error.New("Failed starting node "+newNodeConfig.NodeConfig.Name, err)
			}
			spawner.mutex.Unlock()
			return "", nil
		},
		STOP_ALL_NODES_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			spawner.mutex.Lock()
			defer spawner.mutex.Unlock()
			for _, node := range spawner.nodes {
				err := spawner.stopNode(node.GetName())
				if err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed stopping node "+node.GetName(), err).Error())
					}

				}
			}
			return "", nil
		},
		DESPAWN_ALL_NODES_SYNC: func(node *Node.Node, message *Message.Message) (string, error) {
			spawner.mutex.Lock()
			defer spawner.mutex.Unlock()
			for _, node := range spawner.nodes {
				err := spawner.despawnNode(node.GetName())
				if err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed despawning node "+node.GetName(), err).Error())
					}
				}
			}
			return "", nil
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
				return Error.New("Failed stopping node "+id, err)
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
		STOP_ALL_NODES_ASYNC: func(node *Node.Node, message *Message.Message) error {
			spawner.mutex.Lock()
			defer spawner.mutex.Unlock()
			for _, node := range spawner.nodes {
				err := spawner.stopNode(node.GetName())
				if err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed stopping node "+node.GetName(), err).Error())
					}
				}
			}
			return nil
		},
		DESPAWN_ALL_NODES_ASYNC: func(node *Node.Node, message *Message.Message) error {
			spawner.mutex.Lock()
			defer spawner.mutex.Unlock()
			for _, node := range spawner.nodes {
				err := spawner.despawnNode(node.GetName())
				if err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed despawning node "+node.GetName(), err).Error())
					}
				}
			}
			return nil
		},
	}
}
