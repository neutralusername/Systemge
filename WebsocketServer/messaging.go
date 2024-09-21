package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

// Broadcast broadcasts a message to all connected clients.
// Blocking until all messages are sent.
func (server *WebsocketServer) Broadcast(message *Message.Message) *Event.Event {
	if event := server.onInfo(Event.New(
		Event.SendingMessage,
		server.GetServerContext().Merge(Event.Context{
			"messageType": "websocketBroadcast",
			"topic":       message.GetTopic(),
			"payload":     message.GetPayload(),
			"syncToken":   message.GetSyncToken(),
			"info":        "broadcasting message to all connected clients",
		}),
	)); event.IsError() {
		return event
	}
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.clientMutex.RLock()
	for _, client := range server.clients {
		if !client.isAccepted {
			event := server.onWarning(Event.New(
				Event.ClientNotAccepted,
				server.GetServerContext().Merge(Event.Context{
					"messageType":       "websocketBroadcast",
					"topic":             message.GetTopic(),
					"payload":           message.GetPayload(),
					"syncToken":         message.GetSyncToken(),
					"warning":           "Client is not accepted",
					"targetWebsocketId": client.GetId(),
				}),
			))
			if event.IsError() {
				server.clientMutex.RUnlock()
				return event
			}
			if event.IsWarning() {
				continue
			}
		}
		waitGroup.AddTask(func() {
			server.Send(client, messageBytes)
		})
	}
	server.clientMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()
	return server.onInfo(Event.New(
		Event.SentMessage,
		server.GetServerContext().Merge(Event.Context{
			"messageType": "websocketBroadcast",
			"topic":       message.GetTopic(),
			"payload":     message.GetPayload(),
			"syncToken":   message.GetSyncToken(),
			"info":        "broadcasted message to all connected clients",
		}),
	))
}

// Unicast unicasts a message to a specific client by id.
// Blocking until the message is sent.
func (server *WebsocketServer) Unicast(id string, message *Message.Message) *Event.Event {
	if event := server.onInfo(Event.New(
		Event.SendingMessage,
		server.GetServerContext().Merge(Event.Context{
			"messageType":       "websocketUnicast",
			"topic":             message.GetTopic(),
			"payload":           message.GetPayload(),
			"syncToken":         message.GetSyncToken(),
			"info":              "unicasting message to client",
			"targetWebsocketId": id,
		}),
	)); event.IsError() {
		return event
	}
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.clientMutex.RLock()
	client, exists := server.clients[id]
	if !exists {
		server.clientMutex.RUnlock()
		return server.onError(Event.New(
			Event.ClientDoesNotExist,
			server.GetServerContext().Merge(Event.Context{
				"messageType":       "websocketUnicast",
				"topic":             message.GetTopic(),
				"payload":           message.GetPayload(),
				"syncToken":         message.GetSyncToken(),
				"error":             "Client does not exist",
				"targetWebsocketId": id,
			}),
		))
	}
	if !client.isAccepted {
		event := server.onWarning(Event.New(
			Event.ClientNotAccepted,
			server.GetServerContext().Merge(Event.Context{
				"messageType":       "websocketUnicast",
				"topic":             message.GetTopic(),
				"payload":           message.GetPayload(),
				"syncToken":         message.GetSyncToken(),
				"warning":           "Client is not accepted",
				"targetWebsocketId": id,
			}),
		))
		if event.IsError() {
			server.clientMutex.RUnlock()
			return event
		}
	}
	waitGroup.AddTask(func() {
		server.Send(client, messageBytes)
	})
	server.clientMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()
	return server.onInfo(Event.New(
		Event.SentMessage,
		server.GetServerContext().Merge(Event.Context{
			"messageType":       "websocketUnicast",
			"topic":             message.GetTopic(),
			"payload":           message.GetPayload(),
			"syncToken":         message.GetSyncToken(),
			"info":              "unicasted message to client",
			"targetWebsocketId": id,
		}),
	))
}

// Multicast multicasts a message to multiple clients by id.
// Blocking until all messages are sent.
func (server *WebsocketServer) Multicast(ids []string, message *Message.Message) *Event.Event {
	if event := server.onInfo(Event.New(
		Event.SendingMessage,
		server.GetServerContext().Merge(Event.Context{
			"messageType":        "websocketMulticast",
			"topic":              message.GetTopic(),
			"payload":            message.GetPayload(),
			"syncToken":          message.GetSyncToken(),
			"info":               "multicasting message to clients",
			"targetWebsocketIds": Helpers.JsonMarshal(ids),
		}),
	)); event.IsError() {
		return event
	}
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()
	server.clientMutex.RLock()
	for _, id := range ids {
		client, exists := server.clients[id]
		if !exists {
			if event := server.onError(Event.New(
				Event.ClientDoesNotExist,
				server.GetServerContext().Merge(Event.Context{
					"messageType":        "websocketMulticast",
					"topic":              message.GetTopic(),
					"payload":            message.GetPayload(),
					"syncToken":          message.GetSyncToken(),
					"error":              "Client does not exist",
					"targetWebsocketIds": Helpers.JsonMarshal(ids),
					"targetWebsocketId":  id,
				}),
			)); event.IsError() {
				server.clientMutex.RUnlock()
				return event
			}
		}
		if !client.isAccepted {
			event := server.onWarning(Event.New(
				Event.ClientNotAccepted,
				server.GetServerContext().Merge(Event.Context{
					"messageType":        "websocketMulticast",
					"topic":              message.GetTopic(),
					"payload":            message.GetPayload(),
					"syncToken":          message.GetSyncToken(),
					"warning":            "Client is not accepted",
					"targetWebsocketIds": Helpers.JsonMarshal(ids),
					"targetWebsocketId":  id,
				}),
			))
			if event.IsError() {
				server.clientMutex.RUnlock()
				return event
			}
			if event.IsWarning() {
				continue
			}
		}
		waitGroup.AddTask(func() {
			server.Send(client, messageBytes)
		})
	}
	server.clientMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()
	return server.onInfo(Event.New(
		Event.SentMessage,
		server.GetServerContext().Merge(Event.Context{
			"messageType":        "websocketMulticast",
			"topic":              message.GetTopic(),
			"payload":            message.GetPayload(),
			"syncToken":          message.GetSyncToken(),
			"info":               "multicasted message to clients",
			"targetWebsocketIds": Helpers.JsonMarshal(ids),
		}),
	))
}

// Groupcast groupcasts a message to all clients in a group.
// Blocking until all messages are sent.
func (server *WebsocketServer) Groupcast(groupId string, message *Message.Message) *Event.Event {
	if event := server.onInfo(Event.New(
		Event.SendingMessage,
		server.GetServerContext().Merge(Event.Context{
			"messageType": "websocketGroupcast",
			"topic":       message.GetTopic(),
			"payload":     message.GetPayload(),
			"syncToken":   message.GetSyncToken(),
			"info":        "groupcasting message to group",
			"groupId":     groupId,
		}),
	)); event.IsError() {
		return event
	}
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.clientMutex.RLock()
	group, ok := server.groups[groupId]
	if !ok {
		server.clientMutex.RUnlock()
		return server.onError(Event.New(
			Event.GroupDoesNotExist,
			server.GetServerContext().Merge(Event.Context{
				"messageType": "websocketGroupcast",
				"topic":       message.GetTopic(),
				"payload":     message.GetPayload(),
				"syncToken":   message.GetSyncToken(),
				"error":       "Group does not exist",
				"groupId":     groupId,
			}),
		))
	}
	for _, client := range group {
		if !client.isAccepted {
			event := server.onWarning(Event.New(
				Event.ClientNotAccepted,
				server.GetServerContext().Merge(Event.Context{
					"messageType":       "websocketGroupcast",
					"topic":             message.GetTopic(),
					"payload":           message.GetPayload(),
					"syncToken":         message.GetSyncToken(),
					"warning":           "Client is not accepted",
					"targetWebsocketId": client.GetId(),
					"groupId":           groupId,
				}),
			))
			if event.IsError() {
				server.clientMutex.RUnlock()
				return event
			}
			if event.IsWarning() {
				continue
			}
		}
		waitGroup.AddTask(func() {
			server.Send(client, messageBytes)
		})
	}
	server.clientMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()
	return server.onInfo(Event.New(
		Event.SentMessage,
		server.GetServerContext().Merge(Event.Context{
			"messageType": "websocketGroupcast",
			"topic":       message.GetTopic(),
			"payload":     message.GetPayload(),
			"syncToken":   message.GetSyncToken(),
			"info":        "groupcasted message to group",
			"groupId":     groupId,
		}),
	))
}
