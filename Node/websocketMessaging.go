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
				node.config.Logger.Warning("failed to broadcast message to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\" on node \"" + node.GetName() + "\"")
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
				node.config.Logger.Warning("failed to unicast message to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\" on node \"" + node.GetName() + "\"")
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
					node.config.Logger.Warning("failed to multicast message to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\" on node \"" + node.GetName() + "\"")
				}
			}()
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
		go func() {
			err := websocketClient.Send(messageBytes)
			if err != nil {
				node.config.Logger.Warning("failed to groupcast message to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\" on node \"" + node.GetName() + "\"")
			}
		}()
	}
}
