package Node

import (
	"Systemge/Message"
)

func (client *Node) WebsocketBroadcast(message *Message.Message) {
	messageBytes := message.Serialize()
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	for _, websocketClient := range client.websocketClients {
		go websocketClient.Send(messageBytes)
	}
}

func (client *Node) WebsocketUnicast(id string, message *Message.Message) {
	messageBytes := message.Serialize()
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	if websocketClient, exists := client.websocketClients[id]; exists {
		go websocketClient.Send(messageBytes)
	}
}

func (client *Node) WebsocketMulticast(ids []string, message *Message.Message) {
	messageBytes := message.Serialize()
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	for _, id := range ids {
		if websocketClient, exists := client.websocketClients[id]; exists {
			go websocketClient.Send(messageBytes)
		}
	}
}

func (client *Node) WebsocketGroupcast(groupId string, message *Message.Message) {
	messageBytes := message.Serialize()
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	if client.WebsocketGroups[groupId] == nil {
		return
	}
	for _, websocketClient := range client.WebsocketGroups[groupId] {
		go websocketClient.Send(messageBytes)
	}
}
