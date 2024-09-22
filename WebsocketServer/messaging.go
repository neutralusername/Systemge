package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

// Broadcast broadcasts a message to all connected clients.
// Blocking until all messages are sent.
func (server *WebsocketServer) Broadcast(message *Message.Message) error {
	if event := server.onInfo(Event.New(
		Event.SendingMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":      "broadcasting message to all connected clients",
			"type":      "websocketBroadcast",
			"topic":     message.GetTopic(),
			"payload":   message.GetPayload(),
			"syncToken": message.GetSyncToken(),
			"onError":   "cancel",
			"onWarning": "continue",
			"onInfo":    "continue",
		}),
	)); event.IsError() {
		return event.GetError()
	}

	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.clientMutex.RLock()
	for _, client := range server.clients {
		if !client.isAccepted {
			event := server.onWarning(Event.New(
				Event.ClientNotAccepted,
				server.GetServerContext().Merge(Event.Context{
					"warning":           "Client is not accepted",
					"type":              "websocketBroadcast",
					"targetWebsocketId": client.GetId(),
					"topic":             message.GetTopic(),
					"payload":           message.GetPayload(),
					"syncToken":         message.GetSyncToken(),
					"onError":           "cancel",
					"onWarning":         "skip",
					"onInfo":            "continue",
				}),
			))
			if event.IsError() {
				server.clientMutex.RUnlock()
				return event.GetError()
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

	server.onInfo(Event.New(
		Event.SentMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":      "broadcasted message to all connected clients",
			"type":      "websocketBroadcast",
			"topic":     message.GetTopic(),
			"payload":   message.GetPayload(),
			"syncToken": message.GetSyncToken(),
		}),
	))
	return nil
}

// Unicast unicasts a message to a specific client by id.
// Blocking until the message is sent.
func (server *WebsocketServer) Unicast(id string, message *Message.Message) error {
	if event := server.onInfo(Event.New(
		Event.SendingMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":              "unicasting message to client",
			"type":              "websocketUnicast",
			"targetWebsocketId": id,
			"topic":             message.GetTopic(),
			"payload":           message.GetPayload(),
			"syncToken":         message.GetSyncToken(),
			"onError":           "cancel",
			"onWarning":         "continue",
			"onInfo":            "continue",
		}),
	)); event.IsError() {
		return event.GetError()
	}
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.clientMutex.RLock()
	client, exists := server.clients[id]
	if !exists {
		server.clientMutex.RUnlock()
		server.onError(Event.New(
			Event.ClientDoesNotExist,
			server.GetServerContext().Merge(Event.Context{
				"error":             "Client does not exist",
				"type":              "websocketConnection",
				"targetWebsocketId": id,
				"topic":             message.GetTopic(),
				"payload":           message.GetPayload(),
				"syncToken":         message.GetSyncToken(),
			}),
		))
		return errors.New("client does not exist")
	}
	if !client.isAccepted {
		event := server.onError(Event.New(
			Event.ClientNotAccepted,
			server.GetServerContext().Merge(Event.Context{
				"error":             "Client is not accepted",
				"type":              "websocketUnicast",
				"targetWebsocketId": id,
				"topic":             message.GetTopic(),
				"payload":           message.GetPayload(),
				"syncToken":         message.GetSyncToken(),
				"onError":           "cancel",
				"onWarning":         "continue",
				"onInfo":            "continue",
			}),
		))
		if event.IsError() {
			server.clientMutex.RUnlock()
			return event.GetError()
		}
	}
	waitGroup.AddTask(func() {
		server.Send(client, messageBytes)
	})
	server.clientMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()

	server.onInfo(Event.New(
		Event.SentMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":              "unicasted message to client",
			"type":              "websocketUnicast",
			"targetWebsocketId": id,
			"topic":             message.GetTopic(),
			"payload":           message.GetPayload(),
			"syncToken":         message.GetSyncToken(),
		}),
	))
	return nil
}

// Multicast multicasts a message to multiple clients by id.
// Blocking until all messages are sent.
func (server *WebsocketServer) Multicast(ids []string, message *Message.Message) error {
	if event := server.onInfo(Event.New(
		Event.SendingMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":               "multicasting message to clients",
			"type":               "websocketMulticast",
			"targetWebsocketIds": Helpers.JsonMarshal(ids),
			"topic":              message.GetTopic(),
			"payload":            message.GetPayload(),
			"syncToken":          message.GetSyncToken(),
			"onError":            "cancel",
			"onWarning":          "continue",
			"onInfo":             "continue",
		}),
	)); event.IsError() {
		return event.GetError()
	}
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()
	server.clientMutex.RLock()
	for _, id := range ids {
		client, exists := server.clients[id]
		if !exists {
			event := server.onError(Event.New(
				Event.ClientDoesNotExist,
				server.GetServerContext().Merge(Event.Context{
					"error":              "Client does not exist",
					"type":               "websocketConnection",
					"targetWebsocketId":  id,
					"targetWebsocketIds": Helpers.JsonMarshal(ids),
					"topic":              message.GetTopic(),
					"payload":            message.GetPayload(),
					"syncToken":          message.GetSyncToken(),
					"onError":            "cancel",
					"onWarning":          "skip",
					"onInfo":             "skip",
				}),
			))
			if event.IsError() {
				server.clientMutex.RUnlock()
				return event.GetError()
			} else {
				continue
			}
		}
		if !client.isAccepted {
			event := server.onWarning(Event.New(
				Event.ClientNotAccepted,
				server.GetServerContext().Merge(Event.Context{
					"warning":            "Client is not accepted",
					"type":               "websocketMulticast",
					"targetWebsocketId":  id,
					"targetWebsocketIds": Helpers.JsonMarshal(ids),
					"topic":              message.GetTopic(),
					"payload":            message.GetPayload(),
					"syncToken":          message.GetSyncToken(),
					"onError":            "cancel",
					"onWarning":          "skip",
					"onInfo":             "continue",
				}),
			))
			if event.IsError() {
				server.clientMutex.RUnlock()
				return event.GetError()
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

	server.onInfo(Event.New(
		Event.SentMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":               "multicasted message to clients",
			"type":               "websocketMulticast",
			"targetWebsocketIds": Helpers.JsonMarshal(ids),
			"topic":              message.GetTopic(),
			"payload":            message.GetPayload(),
			"syncToken":          message.GetSyncToken(),
		}),
	))
	return nil
}

// Groupcast groupcasts a message to all clients in a group.
// Blocking until all messages are sent.
func (server *WebsocketServer) Groupcast(groupId string, message *Message.Message) error {
	if event := server.onInfo(Event.New(
		Event.SendingMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":      "groupcasting message to group",
			"type":      "websocketGroupcast",
			"groupId":   groupId,
			"topic":     message.GetTopic(),
			"payload":   message.GetPayload(),
			"syncToken": message.GetSyncToken(),
			"onError":   "cancel",
			"onWarning": "continue",
			"onInfo":    "continue",
		}),
	)); event.IsError() {
		return event.GetError()
	}
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.clientMutex.RLock()
	group, ok := server.groups[groupId]
	if !ok {
		server.clientMutex.RUnlock()
		server.onError(Event.New(
			Event.GroupDoesNotExist,
			server.GetServerContext().Merge(Event.Context{
				"error":     "Group does not exist",
				"type":      "websocketGroupcast",
				"groupId":   groupId,
				"topic":     message.GetTopic(),
				"payload":   message.GetPayload(),
				"syncToken": message.GetSyncToken(),
			}),
		))
		return errors.New("group does not exist")
	}
	for _, client := range group {
		if !client.isAccepted {
			event := server.onWarning(Event.New(
				Event.ClientNotAccepted,
				server.GetServerContext().Merge(Event.Context{
					"warning":           "Client is not accepted",
					"type":              "websocketGroupcast",
					"groupId":           groupId,
					"targetWebsocketId": client.GetId(),
					"topic":             message.GetTopic(),
					"payload":           message.GetPayload(),
					"syncToken":         message.GetSyncToken(),
					"onError":           "cancel",
					"onWarning":         "skip",
					"onInfo":            "continue",
				}),
			))
			if event.IsError() {
				server.clientMutex.RUnlock()
				return event.GetError()
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

	server.onInfo(Event.New(
		Event.SentMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":      "groupcasted message to group",
			"type":      "websocketGroupcast",
			"groupId":   groupId,
			"topic":     message.GetTopic(),
			"payload":   message.GetPayload(),
			"syncToken": message.GetSyncToken(),
		}),
	))
	return nil
}
