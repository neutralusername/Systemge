package Node

import (
	"Systemge/Message"
)

func (node *Node) WebsocketBroadcast(message *Message.Message) {
	if websocket := node.websocket; websocket != nil {
		messageBytes := message.Serialize()
		websocket.mutex.Lock()
		defer websocket.mutex.Unlock()
		for _, websocketClient := range websocket.clients {
			go func() {
				err := websocketClient.Send(messageBytes)
				if err != nil {
					if warningLogger := node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log("Failed to broadcast message with topic \"" + message.GetTopic() + "\" to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\"")
					}
				}
				websocket.websocketOutgoingMessageCounter.Add(1)
			}()
		}
	}
}

func (node *Node) WebsocketUnicast(id string, message *Message.Message) {
	if websocket := node.websocket; websocket != nil {
		messageBytes := message.Serialize()
		websocket.mutex.Lock()
		defer websocket.mutex.Unlock()
		if websocketClient, exists := websocket.clients[id]; exists {
			go func() {
				err := websocketClient.Send(messageBytes)
				if err != nil {
					if warningLogger := node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log("Failed to unicast message with topic \"" + message.GetTopic() + "\" to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\"")
					}
				}
				websocket.websocketOutgoingMessageCounter.Add(1)
			}()
		}
	}
}

func (node *Node) WebsocketMulticast(ids []string, message *Message.Message) {
	if websocket := node.websocket; websocket != nil {
		messageBytes := message.Serialize()
		websocket.mutex.Lock()
		defer websocket.mutex.Unlock()
		for _, id := range ids {
			if websocketClient, exists := websocket.clients[id]; exists {
				go func() {
					err := websocketClient.Send(messageBytes)
					if err != nil {
						if warningLogger := node.GetWarningLogger(); warningLogger != nil {
							warningLogger.Log("Failed to multicast message with topic \"" + message.GetTopic() + "\" to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\"")
						}
					}
					websocket.websocketOutgoingMessageCounter.Add(1)
				}()
			}
		}
	}
}

func (node *Node) WebsocketGroupcast(groupId string, message *Message.Message) {
	if websocket := node.websocket; websocket != nil {
		messageBytes := message.Serialize()
		websocket.mutex.Lock()
		defer websocket.mutex.Unlock()
		if websocket.groups[groupId] == nil {
			return
		}
		for _, websocketClient := range websocket.groups[groupId] {
			go func() {
				err := websocketClient.Send(messageBytes)
				if err != nil {
					if warningLogger := node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log("Failed to groupcast message with topic \"" + message.GetTopic() + "\" to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\"")
					}
				}
				websocket.websocketOutgoingMessageCounter.Add(1)
			}()
		}
	}
}
