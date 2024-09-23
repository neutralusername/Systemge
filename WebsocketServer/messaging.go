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

	server.websocketConnectionMutex.RLock()
	for _, websocketConnection := range server.websocketConnections {
		if !websocketConnection.isAccepted {
			event := server.onWarning(Event.NewWarning(
				Event.ClientNotAccepted,
				"websocketConnection is not accepted",
				Event.Cancel,
				Event.Skip,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.AdditionalKind:    Event.Broadcast,
					Event.TargetWebsocketId: websocketConnection.GetId(),
					Event.Topic:             message.GetTopic(),
					Event.Payload:           message.GetPayload(),
					Event.SyncToken:         message.GetSyncToken(),
				}),
			))
			if event.IsError() {
				server.websocketConnectionMutex.RUnlock()
				return event.GetError()
			}
			if event.IsWarning() {
				continue
			}
		}
		waitGroup.AddTask(func() {
			server.Send(websocketConnection, messageBytes)
		})
	}
	server.websocketConnectionMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()

	server.onInfo(Event.NewInfoNoOption(
		Event.SentMessage,
		"broadcasted websocketConnection message",
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

// Unicast unicasts a message to a specific websocketConnection by id.
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

	server.websocketConnectionMutex.RLock()
	websocketConnection, exists := server.websocketConnections[id]
	if !exists {
		server.websocketConnectionMutex.RUnlock()
		server.onWarning(Event.NewWarningNoOption(
			Event.ClientDoesNotExist,
			"websocketConnection does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:              Event.WebsocketConnection,
				Event.AdditionalKind:    Event.Unicast,
				Event.TargetWebsocketId: id,
				Event.Topic:             message.GetTopic(),
				Event.Payload:           message.GetPayload(),
				Event.SyncToken:         message.GetSyncToken(),
			}),
		))
		return errors.New("websocketConnection does not exist")
	}
	if !websocketConnection.isAccepted {
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
			server.websocketConnectionMutex.RUnlock()
			return event.GetError()
		}
	}
	waitGroup.AddTask(func() {
		server.Send(websocketConnection, messageBytes)
	})
	server.websocketConnectionMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()

	server.onInfo(Event.NewInfoNoOption(
		Event.SentMessage,
		"unicasted websocketConnection message",
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

// Multicast multicasts a message to multiple websocketConnections by id.
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
	server.websocketConnectionMutex.RLock()
	for _, id := range ids {
		websocketConnection, exists := server.websocketConnections[id]
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
				server.websocketConnectionMutex.RUnlock()
				return event.GetError()
			} else {
				continue
			}
		}
		if !websocketConnection.isAccepted {
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
				server.websocketConnectionMutex.RUnlock()
				return event.GetError()
			}
			if event.IsWarning() {
				continue
			}
		}
		waitGroup.AddTask(func() {
			server.Send(websocketConnection, messageBytes)
		})
	}
	server.websocketConnectionMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()

	server.onInfo(Event.NewInfoNoOption(
		Event.SentMessage,
		"multicasted websocketConnection message",
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

// Groupcast groupcasts a message to all websocketConnections in a group.
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

	server.websocketConnectionMutex.RLock()
	group, ok := server.groupsWebsocketConnections[groupId]
	if !ok {
		server.websocketConnectionMutex.RUnlock()
		server.onWarning(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
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
	for _, websocketConnection := range group {
		if !websocketConnection.isAccepted {
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
					Event.TargetWebsocketId: websocketConnection.GetId(),
					Event.Topic:             message.GetTopic(),
					Event.Payload:           message.GetPayload(),
					Event.SyncToken:         message.GetSyncToken(),
				}),
			))
			if event.IsError() {
				server.websocketConnectionMutex.RUnlock()
				return event.GetError()
			}
			if event.IsWarning() {
				continue
			}
		}
		waitGroup.AddTask(func() {
			server.Send(websocketConnection, messageBytes)
		})
	}
	server.websocketConnectionMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()

	server.onInfo(Event.NewInfoNoOption(
		Event.SentMessage,
		"groupcasted websocketConnection message",
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
