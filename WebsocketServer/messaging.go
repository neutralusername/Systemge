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
	if event := server.onInfo(Event.NewInfo(
		Event.SendingMessage,
		"broadcasting websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:           Event.WebsocketConnection,
			Event.AdditionalKind: Event.Broadcast,
			Event.Topic:          message.GetTopic(),
			Event.Payload:        message.GetPayload(),
			Event.SyncToken:      message.GetSyncToken(),
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.clientMutex.RLock()
	for _, client := range server.clients {
		if !client.isAccepted {
			event := server.onWarning(Event.NewWarning(
				Event.ClientNotAccepted,
				"websocketConnection is not accepted",
				Event.Cancel,
				Event.Skip,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.AdditionalKind:    Event.Broadcast,
					Event.TargetWebsocketId: client.GetId(),
					Event.Topic:             message.GetTopic(),
					Event.Payload:           message.GetPayload(),
					Event.SyncToken:         message.GetSyncToken(),
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

	server.onInfo(Event.NewInfo(
		Event.SentMessage,
		"broadcasted websocketConnection message",
		Event.NoOption,
		Event.NoOption,
		Event.NoOption,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:           Event.WebsocketConnection,
			Event.AdditionalKind: Event.Broadcast,
			Event.Topic:          message.GetTopic(),
			Event.Payload:        message.GetPayload(),
			Event.SyncToken:      message.GetSyncToken(),
		}),
	))
	return nil
}

// Unicast unicasts a message to a specific client by id.
// Blocking until the message is sent.
func (server *WebsocketServer) Unicast(id string, message *Message.Message) error {
	if event := server.onInfo(Event.NewInfo(
		Event.SendingMessage,
		"unicasting websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:              Event.WebsocketConnection,
			Event.AdditionalKind:    Event.Unicast,
			Event.TargetWebsocketId: id,
			Event.Topic:             message.GetTopic(),
			Event.Payload:           message.GetPayload(),
			Event.SyncToken:         message.GetSyncToken(),
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.clientMutex.RLock()
	client, exists := server.clients[id]
	if !exists {
		server.clientMutex.RUnlock()
		server.onWarning(Event.NewWarning(
			Event.ClientDoesNotExist,
			"websocketConnection does not exist",
			Event.NoOption,
			Event.NoOption,
			Event.NoOption,
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:              Event.WebsocketConnection,
				Event.AdditionalKind:    Event.Unicast,
				Event.TargetWebsocketId: id,
				Event.Topic:             message.GetTopic(),
				Event.Payload:           message.GetPayload(),
				Event.SyncToken:         message.GetSyncToken(),
			}),
		))
		return errors.New("client does not exist")
	}
	if !client.isAccepted {
		event := server.onWarning(Event.NewWarning(
			Event.ClientNotAccepted,
			"websocketConnection is not accepted",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:              Event.WebsocketConnection,
				Event.AdditionalKind:    Event.Unicast,
				Event.TargetWebsocketId: id,
				Event.Topic:             message.GetTopic(),
				Event.Payload:           message.GetPayload(),
				Event.SyncToken:         message.GetSyncToken(),
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

	server.onInfo(Event.NewInfo(
		Event.SentMessage,
		"unicasted websocketConnection message",
		Event.NoOption,
		Event.NoOption,
		Event.NoOption,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:              Event.WebsocketConnection,
			Event.AdditionalKind:    Event.Unicast,
			Event.TargetWebsocketId: id,
			Event.Topic:             message.GetTopic(),
			Event.Payload:           message.GetPayload(),
			Event.SyncToken:         message.GetSyncToken(),
		}),
	))
	return nil
}

// Multicast multicasts a message to multiple clients by id.
// Blocking until all messages are sent.
func (server *WebsocketServer) Multicast(ids []string, message *Message.Message) error {
	if event := server.onInfo(Event.NewInfo(
		Event.SendingMessage,
		"multicasting websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:               Event.WebsocketConnection,
			Event.AdditionalKind:     Event.Multicast,
			Event.TargetWebsocketIds: Helpers.JsonMarshal(ids),
			Event.Topic:              message.GetTopic(),
			Event.Payload:            message.GetPayload(),
			Event.SyncToken:          message.GetSyncToken(),
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()
	server.clientMutex.RLock()
	for _, id := range ids {
		client, exists := server.clients[id]
		if !exists {
			event := server.onWarning(Event.NewWarning(
				Event.ClientDoesNotExist,
				"websocketConnection does not exist",
				Event.Cancel,
				Event.Skip,
				Event.Skip,
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:               Event.WebsocketConnection,
					Event.AdditionalKind:     Event.Multicast,
					Event.TargetWebsocketId:  id,
					Event.TargetWebsocketIds: Helpers.JsonMarshal(ids),
					Event.Topic:              message.GetTopic(),
					Event.Payload:            message.GetPayload(),
					Event.SyncToken:          message.GetSyncToken(),
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
			event := server.onWarning(Event.NewWarning(
				Event.ClientNotAccepted,
				"websocketConnection is not accepted",
				Event.Cancel,
				Event.Skip,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:               Event.WebsocketConnection,
					Event.AdditionalKind:     Event.Multicast,
					Event.TargetWebsocketId:  id,
					Event.TargetWebsocketIds: Helpers.JsonMarshal(ids),
					Event.Topic:              message.GetTopic(),
					Event.Payload:            message.GetPayload(),
					Event.SyncToken:          message.GetSyncToken(),
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

	server.onInfo(Event.NewInfo(
		Event.SentMessage,
		"multicasted websocketConnection message",
		Event.NoOption,
		Event.NoOption,
		Event.NoOption,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:               Event.WebsocketConnection,
			Event.TargetWebsocketIds: Helpers.JsonMarshal(ids),
			Event.Topic:              message.GetTopic(),
			Event.Payload:            message.GetPayload(),
			Event.SyncToken:          message.GetSyncToken(),
		}),
	))
	return nil
}

// Groupcast groupcasts a message to all clients in a group.
// Blocking until all messages are sent.
func (server *WebsocketServer) Groupcast(groupId string, message *Message.Message) error {
	if event := server.onInfo(Event.NewInfo(
		Event.SendingMessage,
		"groupcasting websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:           Event.WebsocketConnection,
			Event.AdditionalKind: Event.Groupcast,
			Event.GroupId:        groupId,
			Event.Topic:          message.GetTopic(),
			Event.Payload:        message.GetPayload(),
			Event.SyncToken:      message.GetSyncToken(),
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.clientMutex.RLock()
	group, ok := server.groups[groupId]
	if !ok {
		server.clientMutex.RUnlock()
		server.onWarning(Event.NewWarning(
			Event.GroupDoesNotExist,
			"group does not exist",
			Event.NoOption,
			Event.NoOption,
			Event.NoOption,
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:           Event.WebsocketConnection,
				Event.AdditionalKind: Event.Groupcast,
				Event.GroupId:        groupId,
				Event.Topic:          message.GetTopic(),
				Event.Payload:        message.GetPayload(),
				Event.SyncToken:      message.GetSyncToken(),
			}),
		))
		return errors.New("group does not exist")
	}
	for _, client := range group {
		if !client.isAccepted {
			event := server.onWarning(Event.NewWarning(
				Event.ClientNotAccepted,
				"websocketConnection is not accepted",
				Event.Cancel,
				Event.Skip,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.AdditionalKind:    Event.Groupcast,
					Event.GroupId:           groupId,
					Event.TargetWebsocketId: client.GetId(),
					Event.Topic:             message.GetTopic(),
					Event.Payload:           message.GetPayload(),
					Event.SyncToken:         message.GetSyncToken(),
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

	server.onInfo(Event.NewInfo(
		Event.SentMessage,
		"groupcasted websocketConnection message",
		Event.NoOption,
		Event.NoOption,
		Event.NoOption,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:           Event.WebsocketConnection,
			Event.AdditionalKind: Event.Groupcast,
			Event.GroupId:        groupId,
			Event.Topic:          message.GetTopic(),
			Event.Payload:        message.GetPayload(),
			Event.SyncToken:      message.GetSyncToken(),
		}),
	))
	return nil
}
