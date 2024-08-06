package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

// WebsocketBroadcast broadcasts a message to all connected websocket clients.
// Blocking until all messages are sent.
func (node *Node) WebsocketBroadcast(message *Message.Message) error {
	if websocket := node.websocket; websocket != nil {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log("Broadcasting message with topic \"" + message.GetTopic() + "\"")
		}
		messageBytes := message.Serialize()
		waitGroup := Tools.NewWaitgroup()
		websocket.mutex.RLock()
		for _, websocketClient := range websocket.clients {
			waitGroup.Add(func() {
				err := websocketClient.Send(messageBytes)
				if err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log("Failed to broadcast message with topic \"" + message.GetTopic() + "\" to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\"")
					}
					if mailer := node.GetMailer(); mailer != nil {
						err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to broadcast message with topic \""+message.GetTopic()+"\" to websocketClient \""+websocketClient.GetId()+"\" with ip \""+websocketClient.GetIp()+"\"", err).Error()))
						if err != nil {
							if errorLogger := node.GetErrorLogger(); errorLogger != nil {
								errorLogger.Log(Error.New("Failed sending mail", err).Error())
							}
						}
					}
				}
				websocket.outgoigMessageCounter.Add(1)
				websocket.bytesSentCounter.Add(uint64(len(messageBytes)))
			})
		}
		websocket.mutex.RUnlock()
		waitGroup.Execute()
		return nil
	}
	return Error.New("Websocket is not initialized", nil)
}

// WebsocketUnicast unicasts a message to a specific websocket client by id.
// Blocking until the message is sent.
func (node *Node) WebsocketUnicast(id string, message *Message.Message) error {
	if websocket := node.websocket; websocket != nil {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log("Unicasting message with topic \"" + message.GetTopic() + "\" to websocketClient \"" + id + "\"")
		}
		messageBytes := message.Serialize()
		waitGroup := Tools.NewWaitgroup()
		websocket.mutex.RLock()
		if websocketClient, exists := websocket.clients[id]; exists {
			waitGroup.Add(func() {
				err := websocketClient.Send(messageBytes)
				if err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log("Failed to unicast message with topic \"" + message.GetTopic() + "\" to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\"")
					}
					if mailer := node.GetMailer(); mailer != nil {
						err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to unicast message with topic \""+message.GetTopic()+"\" to websocketClient \""+websocketClient.GetId()+"\" with ip \""+websocketClient.GetIp()+"\"", err).Error()))
						if err != nil {
							if errorLogger := node.GetErrorLogger(); errorLogger != nil {
								errorLogger.Log(Error.New("Failed sending mail", err).Error())
							}
						}
					}
				}
				websocket.outgoigMessageCounter.Add(1)
				websocket.bytesSentCounter.Add(uint64(len(messageBytes)))
			})
		}
		websocket.mutex.RUnlock()
		waitGroup.Execute()
		return nil
	}
	return Error.New("Websocket is not initialized", nil)
}

// WebsocketMulticast multicasts a message to multiple websocket clients by id.
// Blocking until all messages are sent.
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
		websocket.mutex.RLock()
		for _, id := range ids {
			if websocketClient, exists := websocket.clients[id]; exists {
				waitGroup.Add(func() {
					err := websocketClient.Send(messageBytes)
					if err != nil {
						if errorLogger := node.GetErrorLogger(); errorLogger != nil {
							errorLogger.Log("Failed to multicast message with topic \"" + message.GetTopic() + "\" to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\"")
						}
						if mailer := node.GetMailer(); mailer != nil {
							err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to multicast message with topic \""+message.GetTopic()+"\" to websocketClient \""+websocketClient.GetId()+"\" with ip \""+websocketClient.GetIp()+"\"", err).Error()))
							if err != nil {
								if errorLogger := node.GetErrorLogger(); errorLogger != nil {
									errorLogger.Log(Error.New("Failed sending mail", err).Error())
								}
							}
						}
					}
					websocket.outgoigMessageCounter.Add(1)
					websocket.bytesSentCounter.Add(uint64(len(messageBytes)))
				})
			}
		}
		websocket.mutex.RUnlock()
		waitGroup.Execute()
		return nil
	}
	return Error.New("Websocket is not initialized", nil)
}

// WebsocketGroupcast groupcasts a message to all websocket clients in a group.
// Blocking until all messages are sent.
func (node *Node) WebsocketGroupcast(groupId string, message *Message.Message) error {
	if websocket := node.websocket; websocket != nil {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log("Groupcasting message with topic \"" + message.GetTopic() + "\" to group \"" + groupId + "\"")
		}
		messageBytes := message.Serialize()
		waitGroup := Tools.NewWaitgroup()
		websocket.mutex.RLock()
		if websocket.groups[groupId] == nil {
			websocket.mutex.RUnlock()
			return Error.New("Group \""+groupId+"\" does not exist", nil)
		}
		for _, websocketClient := range websocket.groups[groupId] {
			waitGroup.Add(func() {
				err := websocketClient.Send(messageBytes)
				if err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log("Failed to groupcast message with topic \"" + message.GetTopic() + "\" to websocketClient \"" + websocketClient.GetId() + "\" with ip \"" + websocketClient.GetIp() + "\"")
					}
					if mailer := node.GetMailer(); mailer != nil {
						err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to groupcast message with topic \""+message.GetTopic()+"\" to websocketClient \""+websocketClient.GetId()+"\" with ip \""+websocketClient.GetIp()+"\"", err).Error()))
						if err != nil {
							if errorLogger := node.GetErrorLogger(); errorLogger != nil {
								errorLogger.Log(Error.New("Failed sending mail", err).Error())
							}
						}
					}
				}
				websocket.outgoigMessageCounter.Add(1)
				websocket.bytesSentCounter.Add(uint64(len(messageBytes)))
			})
		}
		websocket.mutex.RUnlock()
		waitGroup.Execute()
		return nil
	}
	return Error.New("Websocket is not initialized", nil)
}
