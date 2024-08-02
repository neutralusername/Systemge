package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (node *Node) WebsocketBroadcast(message *Message.Message) error {
	if websocket := node.websocket; websocket != nil {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log("Broadcasting message with topic \"" + message.GetTopic() + "\"")
		}
		messageBytes := message.Serialize()
		waitGroup := Tools.NewWaitgroup()
		websocket.mutex.Lock()
		for _, websocketClient := range websocket.clients {
			waitGroup.Add(func() {
				err := websocketClient.Send(messageBytes)
				if err != nil {
					if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
						warningLogger.Log("Failed to broadcast message with topic \"" + message.GetTopic() + "\" to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\"")
					}
				}
				websocket.outgoigMessageCounter.Add(1)
				websocket.bytesSentCounter.Add(uint64(len(messageBytes)))
			})
		}
		websocket.mutex.Unlock()
		waitGroup.Execute()
		return nil
	}
	return Error.New("Websocket is not initialized", nil)
}

func (node *Node) WebsocketUnicast(id string, message *Message.Message) error {
	if websocket := node.websocket; websocket != nil {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log("Unicasting message with topic \"" + message.GetTopic() + "\" to websocketClient \"" + id + "\"")
		}
		messageBytes := message.Serialize()
		waitGroup := Tools.NewWaitgroup()
		websocket.mutex.Lock()
		if websocketClient, exists := websocket.clients[id]; exists {
			waitGroup.Add(func() {
				err := websocketClient.Send(messageBytes)
				if err != nil {
					if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
						warningLogger.Log("Failed to unicast message with topic \"" + message.GetTopic() + "\" to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\"")
					}
				}
				websocket.outgoigMessageCounter.Add(1)
				websocket.bytesSentCounter.Add(uint64(len(messageBytes)))
			})
		}
		websocket.mutex.Unlock()
		waitGroup.Execute()
		return nil
	}
	return Error.New("Websocket is not initialized", nil)
}

func (node *Node) WebsocketMulticast(ids []string, message *Message.Message) error {
	if websocket := node.websocket; websocket != nil {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			idsString := ""
			for _, id := range ids {
				idsString += id + ", "
			}
			infoLogger.Log("Multicasting message with topic \"" + message.GetTopic() + "\" to websocketClients \"" + idsString[:len(idsString)-2] + "\"")
		}
		messageBytes := message.Serialize()
		waitGroup := Tools.NewWaitgroup()
		websocket.mutex.Lock()
		for _, id := range ids {
			if websocketClient, exists := websocket.clients[id]; exists {
				waitGroup.Add(func() {
					err := websocketClient.Send(messageBytes)
					if err != nil {
						if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
							warningLogger.Log("Failed to multicast message with topic \"" + message.GetTopic() + "\" to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\"")
						}
					}
					websocket.outgoigMessageCounter.Add(1)
					websocket.bytesSentCounter.Add(uint64(len(messageBytes)))
				})
			}
		}
		websocket.mutex.Unlock()
		waitGroup.Execute()
		return nil
	}
	return Error.New("Websocket is not initialized", nil)
}

func (node *Node) WebsocketGroupcast(groupId string, message *Message.Message) error {
	if websocket := node.websocket; websocket != nil {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log("Groupcasting message with topic \"" + message.GetTopic() + "\" to group \"" + groupId + "\"")
		}
		messageBytes := message.Serialize()
		waitGroup := Tools.NewWaitgroup()
		websocket.mutex.Lock()
		if websocket.groups[groupId] == nil {
			websocket.mutex.Unlock()
			return Error.New("Group \""+groupId+"\" does not exist", nil)
		}
		for _, websocketClient := range websocket.groups[groupId] {
			waitGroup.Add(func() {
				err := websocketClient.Send(messageBytes)
				if err != nil {
					if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
						warningLogger.Log("Failed to groupcast message with topic \"" + message.GetTopic() + "\" to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\"")
					}
				}
				websocket.outgoigMessageCounter.Add(1)
				websocket.bytesSentCounter.Add(uint64(len(messageBytes)))
			})
		}
		websocket.mutex.Unlock()
		waitGroup.Execute()
		return nil
	}
	return Error.New("Websocket is not initialized", nil)
}
