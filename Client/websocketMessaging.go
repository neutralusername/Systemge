package Client

import (
	"Systemge/Message"
)

func (client *Client) WebsocketBroadcast(message *Message.Message) {
	messageBytes := message.Serialize()
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	for _, websocketClient := range client.websocketClients {
		go websocketClient.Send(messageBytes)
	}
}

func (client *Client) WebsocketUnicast(id string, message *Message.Message) {
	messageBytes := message.Serialize()
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	if websocketClient, exists := client.websocketClients[id]; exists {
		go websocketClient.Send(messageBytes)
	}
}

func (client *Client) WebsocketMulticast(ids []string, message *Message.Message) {
	messageBytes := message.Serialize()
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	for _, id := range ids {
		if websocketClient, exists := client.websocketClients[id]; exists {
			go websocketClient.Send(messageBytes)
		}
	}
}

func (client *Client) WebsocketGroupcast(groupId string, message *Message.Message) {
	messageBytes := message.Serialize()
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	if client.groups[groupId] == nil {
		return
	}
	for _, websocketClient := range client.groups[groupId] {
		go websocketClient.Send(messageBytes)
	}
}
