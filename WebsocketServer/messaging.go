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
		"broadcasting message to all connected clients",
		Event.Info,
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			"type":      "websocketBroadcast",
			"topic":     message.GetTopic(),
			"payload":   message.GetPayload(),
			"syncToken": message.GetSyncToken(),
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
				"client is not accepted",
				Event.Warning,
				Event.Cancel,
				Event.Skip,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					"type":              "websocketBroadcast",
					"targetWebsocketId": client.GetId(),
					"topic":             message.GetTopic(),
					"payload":           message.GetPayload(),
					"syncToken":         message.GetSyncToken(),
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
		"broadcasted message to all connected clients",
		Event.Info,
		Event.NoOption,
		Event.NoOption,
		Event.NoOption,
		server.GetServerContext().Merge(Event.Context{
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
		"unicasting message to client",
		Event.Info,
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			"type":              "websocketUnicast",
			"targetWebsocketId": id,
			"topic":             message.GetTopic(),
			"payload":           message.GetPayload(),
			"syncToken":         message.GetSyncToken(),
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
		server.onWarning(Event.New(
			Event.ClientDoesNotExist,
			"client does not exist",
			Event.Warning,
			Event.NoOption,
			Event.NoOption,
			Event.NoOption,
			server.GetServerContext().Merge(Event.Context{
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
		event := server.onWarning(Event.New(
			Event.ClientNotAccepted,
			"client is not accepted",
			Event.Warning,
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				"type":              "websocketUnicast",
				"targetWebsocketId": id,
				"topic":             message.GetTopic(),
				"payload":           message.GetPayload(),
				"syncToken":         message.GetSyncToken(),
			}),
		))
		if !event.IsInfo() {
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
		"unicasted message to client",
		Event.Info,
		Event.NoOption,
		Event.NoOption,
		Event.NoOption,
		server.GetServerContext().Merge(Event.Context{
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
		"multicasting message to clients",
		Event.Info,
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			"type":               "websocketMulticast",
			"targetWebsocketIds": Helpers.JsonMarshal(ids),
			"topic":              message.GetTopic(),
			"payload":            message.GetPayload(),
			"syncToken":          message.GetSyncToken(),
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
			event := server.onWarning(Event.New(
				Event.ClientDoesNotExist,
				"client does not exist",
				Event.Warning,
				Event.Cancel,
				Event.Skip,
				Event.Skip,
				server.GetServerContext().Merge(Event.Context{
					"type":               "websocketConnection",
					"targetWebsocketId":  id,
					"targetWebsocketIds": Helpers.JsonMarshal(ids),
					"topic":              message.GetTopic(),
					"payload":            message.GetPayload(),
					"syncToken":          message.GetSyncToken(),
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
				"client is not accepted",
				Event.Warning,
				Event.Cancel,
				Event.Skip,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					"type":               "websocketMulticast",
					"targetWebsocketId":  id,
					"targetWebsocketIds": Helpers.JsonMarshal(ids),
					"topic":              message.GetTopic(),
					"payload":            message.GetPayload(),
					"syncToken":          message.GetSyncToken(),
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
		"multicasted message to clients",
		Event.Info,
		Event.NoOption,
		Event.NoOption,
		Event.NoOption,
		server.GetServerContext().Merge(Event.Context{
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
		"groupcasting message to group",
		Event.Info,
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			"type":      "websocketGroupcast",
			"groupId":   groupId,
			"topic":     message.GetTopic(),
			"payload":   message.GetPayload(),
			"syncToken": message.GetSyncToken(),
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
		server.onWarning(Event.New(
			Event.GroupDoesNotExist,
			"group does not exist",
			Event.Warning,
			Event.NoOption,
			Event.NoOption,
			Event.NoOption,
			server.GetServerContext().Merge(Event.Context{
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
				"client is not accepted",
				Event.Warning,
				Event.Cancel,
				Event.Skip,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					"type":              "websocketGroupcast",
					"groupId":           groupId,
					"targetWebsocketId": client.GetId(),
					"topic":             message.GetTopic(),
					"payload":           message.GetPayload(),
					"syncToken":         message.GetSyncToken(),
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
		"groupcasted message to group",
		Event.Info,
		Event.NoOption,
		Event.NoOption,
		Event.NoOption,
		server.GetServerContext().Merge(Event.Context{
			"type":      "websocketGroupcast",
			"groupId":   groupId,
			"topic":     message.GetTopic(),
			"payload":   message.GetPayload(),
			"syncToken": message.GetSyncToken(),
		}),
	))
	return nil
}
