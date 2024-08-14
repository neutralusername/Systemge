package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

// Broadcast broadcasts a message to all connected websocket clients.
// Blocking until all messages are sent.
func (server *WebsocketServer) Broadcast(message *Message.Message) error {
	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log("Broadcasting message with topic \"" + message.GetTopic() + "\"")
	}
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()
	server.mutex.RLock()
	for _, client := range server.clients {
		waitGroup.AddTask(func() {
			err := client.Send(messageBytes)
			if err != nil {
				if errorLogger := server.errorLogger; errorLogger != nil {
					errorLogger.Log("Failed to broadcast message with topic \"" + message.GetTopic() + "\" to client \"" + client.GetId() + "\" with ip \"" + client.GetIp() + "\"")
				}
				if mailer := server.mailer; mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to broadcast message with topic \""+message.GetTopic()+"\" to client \""+client.GetId()+"\" with ip \""+client.GetIp()+"\"", err).Error()))
					if err != nil {
						if errorLogger := server.errorLogger; errorLogger != nil {
							errorLogger.Log(Error.New("Failed sending mail", err).Error())
						}
					}
				}
			}
			server.outgoigMessageCounter.Add(1)
			server.bytesSentCounter.Add(uint64(len(messageBytes)))
		})
	}
	server.mutex.RUnlock()
	waitGroup.ExecuteTasks()
	return nil
}

// Unicast unicasts a message to a specific websocket client by id.
// Blocking until the message is sent.
func (server *WebsocketServer) Unicast(id string, message *Message.Message) error {
	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log("Unicasting message with topic \"" + message.GetTopic() + "\" to client \"" + id + "\"")
	}
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()
	server.mutex.RLock()
	if client, exists := server.clients[id]; exists {
		waitGroup.AddTask(func() {
			err := client.Send(messageBytes)
			if err != nil {
				if errorLogger := server.errorLogger; errorLogger != nil {
					errorLogger.Log("Failed to unicast message with topic \"" + message.GetTopic() + "\" to client \"" + client.GetId() + "\" with ip \"" + client.GetIp() + "\"")
				}
				if mailer := server.mailer; mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to unicast message with topic \""+message.GetTopic()+"\" to client \""+client.GetId()+"\" with ip \""+client.GetIp()+"\"", err).Error()))
					if err != nil {
						if errorLogger := server.errorLogger; errorLogger != nil {
							errorLogger.Log(Error.New("Failed sending mail", err).Error())
						}
					}
				}
			}
			server.outgoigMessageCounter.Add(1)
			server.bytesSentCounter.Add(uint64(len(messageBytes)))
		})
	}
	server.mutex.RUnlock()
	waitGroup.ExecuteTasks()
	return nil
}

// Multicast multicasts a message to multiple websocket clients by id.
// Blocking until all messages are sent.
func (server *WebsocketServer) Multicast(ids []string, message *Message.Message) error {
	if infoLogger := server.infoLogger; infoLogger != nil {
		idsString := ""
		for _, id := range ids {
			idsString += id + ", "
		}
		infoLogger.Log("Multicasting message with topic \"" + message.GetTopic() + "\" to client \"" + idsString[:len(idsString)-2] + "\"")
	}
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()
	server.mutex.RLock()
	for _, id := range ids {
		if client, exists := server.clients[id]; exists {
			waitGroup.AddTask(func() {
				err := client.Send(messageBytes)
				if err != nil {
					if errorLogger := server.errorLogger; errorLogger != nil {
						errorLogger.Log("Failed to multicast message with topic \"" + message.GetTopic() + "\" to client \"" + client.GetId() + "\" with ip \"" + client.GetIp() + "\"")
					}
					if mailer := server.mailer; mailer != nil {
						err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to multicast message with topic \""+message.GetTopic()+"\" to client \""+client.GetId()+"\" with ip \""+client.GetIp()+"\"", err).Error()))
						if err != nil {
							if errorLogger := server.errorLogger; errorLogger != nil {
								errorLogger.Log(Error.New("Failed sending mail", err).Error())
							}
						}
					}
				}
				server.outgoigMessageCounter.Add(1)
				server.bytesSentCounter.Add(uint64(len(messageBytes)))
			})
		}
	}
	server.mutex.RUnlock()
	waitGroup.ExecuteTasks()
	return nil
}

// Groupcast groupcasts a message to all websocket clients in a group.
// Blocking until all messages are sent.
func (server *WebsocketServer) Groupcast(groupId string, message *Message.Message) error {
	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log("Groupcasting message with topic \"" + message.GetTopic() + "\" to group \"" + groupId + "\"")
	}
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()
	server.mutex.RLock()
	if server.groups[groupId] == nil {
		server.mutex.RUnlock()
		return Error.New("Group \""+groupId+"\" does not exist", nil)
	}
	for _, client := range server.groups[groupId] {
		waitGroup.AddTask(func() {
			err := client.Send(messageBytes)
			if err != nil {
				if errorLogger := server.errorLogger; errorLogger != nil {
					errorLogger.Log("Failed to groupcast message with topic \"" + message.GetTopic() + "\" to client \"" + client.GetId() + "\" with ip \"" + client.GetIp() + "\"")
				}
				if mailer := server.mailer; mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to groupcast message with topic \""+message.GetTopic()+"\" to client \""+client.GetId()+"\" with ip \""+client.GetIp()+"\"", err).Error()))
					if err != nil {
						if errorLogger := server.errorLogger; errorLogger != nil {
							errorLogger.Log(Error.New("Failed sending mail", err).Error())
						}
					}
				}
			}
			server.outgoigMessageCounter.Add(1)
			server.bytesSentCounter.Add(uint64(len(messageBytes)))
		})
	}
	server.mutex.RUnlock()
	waitGroup.ExecuteTasks()
	return nil
}
