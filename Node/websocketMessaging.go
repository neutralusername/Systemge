package Node

import (
	"Systemge/Message"
)

func (node *Node) WebsocketBroadcast(message *Message.Message) {
	messageBytes := message.Serialize()
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	for _, websocketClient := range node.websocketClients {
		go func() {
			err := websocketClient.Send(messageBytes)
			if err != nil {
				if warningLogger := node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log("Failed to broadcast message to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\" on node \"" + node.GetName() + "\"")
				}
			}
		}()
	}
}

func (node *Node) WebsocketUnicast(id string, message *Message.Message) {
	messageBytes := message.Serialize()
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	if websocketClient, exists := node.websocketClients[id]; exists {
		go func() {
			err := websocketClient.Send(messageBytes)
			if err != nil {
				if warningLogger := node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log("Failed to unicast message to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\" on node \"" + node.GetName() + "\"")
				}
			}
		}()
	}
}

func (node *Node) WebsocketMulticast(ids []string, message *Message.Message) {
	messageBytes := message.Serialize()
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	for _, id := range ids {
		if websocketClient, exists := node.websocketClients[id]; exists {
			go func() {
				err := websocketClient.Send(messageBytes)
				if err != nil {
					if warningLogger := node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log("Failed to multicast message to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\" on node \"" + node.GetName() + "\"")
					}
				}
			}()
		}
	}
}

func (node *Node) WebsocketGroupcast(groupId string, message *Message.Message) {
	messageBytes := message.Serialize()
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	if node.websocketGroups[groupId] == nil {
		return
	}
	for _, websocketClient := range node.websocketGroups[groupId] {
		go func() {
			err := websocketClient.Send(messageBytes)
			if err != nil {
				if warningLogger := node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log("Failed to groupcast message to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\" on node \"" + node.GetName() + "\"")
				}
			}
		}()
	}
}
