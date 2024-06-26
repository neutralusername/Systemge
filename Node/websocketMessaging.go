package Node

import (
	"Systemge/Message"
)

func (node *Node) WebsocketBroadcast(message *Message.Message) {
	messageBytes := message.Serialize()
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	for _, websocketClient := range node.websocketClients {
		go websocketClient.Send(messageBytes)
	}
}

func (node *Node) WebsocketUnicast(id string, message *Message.Message) {
	messageBytes := message.Serialize()
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	if websocketClient, exists := node.websocketClients[id]; exists {
		go websocketClient.Send(messageBytes)
	}
}

func (node *Node) WebsocketMulticast(ids []string, message *Message.Message) {
	messageBytes := message.Serialize()
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	for _, id := range ids {
		if websocketClient, exists := node.websocketClients[id]; exists {
			go websocketClient.Send(messageBytes)
		}
	}
}

func (node *Node) WebsocketGroupcast(groupId string, message *Message.Message) {
	messageBytes := message.Serialize()
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	if node.WebsocketGroups[groupId] == nil {
		return
	}
	for _, websocketClient := range node.WebsocketGroups[groupId] {
		go websocketClient.Send(messageBytes)
	}
}
