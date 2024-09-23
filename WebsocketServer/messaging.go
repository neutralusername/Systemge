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
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.websocketConnectionMutex.RLock()
	targetClientIds := Helpers.JsonMarshal(server.GetClientIds())
	if event := server.onInfo(Event.NewInfo(
		Event.SendingMessage,
		"broadcasting websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:    Event.BroadcastRoutine,
			Event.ClientType:      Event.WebsocketConnection,
			Event.TargetClientIds: targetClientIds,
			Event.Topic:           message.GetTopic(),
			Event.Payload:         message.GetPayload(),
			Event.SyncToken:       message.GetSyncToken(),
		}),
	)); !event.IsInfo() {
		server.websocketConnectionMutex.RUnlock()
		return event.GetError()
	}

	for _, websocketConnection := range server.websocketConnections {
		if !websocketConnection.isAccepted {
			event := server.onWarning(Event.NewWarning(
				Event.ClientNotAccepted,
				"websocketConnection is not accepted",
				Event.Cancel,
				Event.Skip,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Circumstance:    Event.BroadcastRoutine,
					Event.ClientType:      Event.WebsocketConnection,
					Event.ClientId:        websocketConnection.GetId(),
					Event.TargetClientIds: targetClientIds,
					Event.Topic:           message.GetTopic(),
					Event.Payload:         message.GetPayload(),
					Event.SyncToken:       message.GetSyncToken(),
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
			server.send(websocketConnection, messageBytes, Event.BroadcastRoutine)
		})
	}
	server.websocketConnectionMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()

	server.onInfo(Event.NewInfoNoOption(
		Event.SentMessage,
		"broadcasted websocketConnection message",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:    Event.BroadcastRoutine,
			Event.ClientType:      Event.WebsocketConnection,
			Event.TargetClientIds: targetClientIds,
			Event.Topic:           message.GetTopic(),
			Event.Payload:         message.GetPayload(),
			Event.SyncToken:       message.GetSyncToken(),
		}),
	))
	return nil
}

// Unicast unicasts a message to a specific websocketConnection by id.
// Blocking until the message is sent.
func (server *WebsocketServer) Unicast(websocketId string, message *Message.Message) error {
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.websocketConnectionMutex.RLock()
	if event := server.onInfo(Event.NewInfo(
		Event.SendingMessage,
		"unicasting websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:   Event.UnicastRoutine,
			Event.ClientType:     Event.WebsocketConnection,
			Event.TargetClientId: websocketId,
			Event.Topic:          message.GetTopic(),
			Event.Payload:        message.GetPayload(),
			Event.SyncToken:      message.GetSyncToken(),
		}),
	)); !event.IsInfo() {
		server.websocketConnectionMutex.RUnlock()
		return event.GetError()
	}

	websocketConnection, exists := server.websocketConnections[websocketId]
	if !exists {
		server.websocketConnectionMutex.RUnlock()
		server.onWarning(Event.NewWarningNoOption(
			Event.ClientDoesNotExist,
			"websocketConnection does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:   Event.UnicastRoutine,
				Event.ClientType:     Event.WebsocketConnection,
				Event.ClientId:       websocketId,
				Event.TargetClientId: websocketId,
				Event.Topic:          message.GetTopic(),
				Event.Payload:        message.GetPayload(),
				Event.SyncToken:      message.GetSyncToken(),
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
				Event.Circumstance:   Event.UnicastRoutine,
				Event.ClientType:     Event.WebsocketConnection,
				Event.ClientId:       websocketId,
				Event.TargetClientId: websocketId,
				Event.Topic:          message.GetTopic(),
				Event.Payload:        message.GetPayload(),
				Event.SyncToken:      message.GetSyncToken(),
			}),
		))
		if !event.IsInfo() {
			server.websocketConnectionMutex.RUnlock()
			return event.GetError()
		}
	}
	waitGroup.AddTask(func() {
		server.send(websocketConnection, messageBytes, Event.UnicastRoutine)
	})
	server.websocketConnectionMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()

	server.onInfo(Event.NewInfoNoOption(
		Event.SentMessage,
		"unicasted websocketConnection message",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:   Event.UnicastRoutine,
			Event.ClientType:     Event.WebsocketConnection,
			Event.TargetClientId: websocketId,
			Event.Topic:          message.GetTopic(),
			Event.Payload:        message.GetPayload(),
			Event.SyncToken:      message.GetSyncToken(),
		}),
	))
	return nil
}

// Multicast multicasts a message to multiple websocketConnections by id.
// Blocking until all messages are sent.
func (server *WebsocketServer) Multicast(ids []string, message *Message.Message) error {
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.websocketConnectionMutex.RLock()
	targetClientIds := Helpers.JsonMarshal(ids)
	if event := server.onInfo(Event.NewInfo(
		Event.SendingMessage,
		"multicasting websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:    Event.MulticastRoutine,
			Event.ClientType:      Event.WebsocketConnection,
			Event.TargetClientIds: targetClientIds,
			Event.Topic:           message.GetTopic(),
			Event.Payload:         message.GetPayload(),
			Event.SyncToken:       message.GetSyncToken(),
		}),
	)); !event.IsInfo() {
		server.websocketConnectionMutex.RUnlock()
		return event.GetError()
	}

	for _, id := range ids {
		websocketConnection, exists := server.websocketConnections[id]
		if !exists {
			event := server.onWarning(Event.NewWarning(
				Event.ClientDoesNotExist,
				"websocketConnection does not exist",
				Event.Cancel,
				Event.Cancel,
				Event.Skip,
				server.GetServerContext().Merge(Event.Context{
					Event.Circumstance:    Event.MulticastRoutine,
					Event.ClientType:      Event.WebsocketConnection,
					Event.ClientId:        id,
					Event.TargetClientIds: targetClientIds,
					Event.Topic:           message.GetTopic(),
					Event.Payload:         message.GetPayload(),
					Event.SyncToken:       message.GetSyncToken(),
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
					Event.Circumstance:    Event.MulticastRoutine,
					Event.ClientType:      Event.WebsocketConnection,
					Event.ClientId:        id,
					Event.TargetClientIds: targetClientIds,
					Event.Topic:           message.GetTopic(),
					Event.Payload:         message.GetPayload(),
					Event.SyncToken:       message.GetSyncToken(),
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
			server.send(websocketConnection, messageBytes, Event.MulticastRoutine)
		})
	}
	server.websocketConnectionMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()

	server.onInfo(Event.NewInfoNoOption(
		Event.SentMessage,
		"multicasted websocketConnection message",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:    Event.MulticastRoutine,
			Event.ClientType:      Event.WebsocketConnection,
			Event.TargetClientIds: targetClientIds,
			Event.Topic:           message.GetTopic(),
			Event.Payload:         message.GetPayload(),
			Event.SyncToken:       message.GetSyncToken(),
		}),
	))
	return nil
}

// Groupcast groupcasts a message to all websocketConnections in a group.
// Blocking until all messages are sent.
func (server *WebsocketServer) Groupcast(groupId string, message *Message.Message) error {
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.websocketConnectionMutex.RLock()
	if event := server.onInfo(Event.NewInfo(
		Event.SendingMessage,
		"groupcasting websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.GroupcastRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.GroupId:      groupId,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		}),
	)); !event.IsInfo() {
		server.websocketConnectionMutex.RUnlock()
		return event.GetError()
	}

	group, ok := server.groupsWebsocketConnections[groupId]
	if !ok {
		server.websocketConnectionMutex.RUnlock()
		server.onWarning(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance: Event.GroupcastRoutine,
				Event.ClientType:   Event.WebsocketConnection,
				Event.GroupId:      groupId,
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
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
					Event.Circumstance: Event.GroupcastRoutine,
					Event.ClientType:   Event.WebsocketConnection,
					Event.ClientId:     websocketConnection.GetId(),
					Event.GroupId:      groupId,
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
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
			server.send(websocketConnection, messageBytes, Event.GroupcastRoutine)
		})
	}
	server.websocketConnectionMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()

	server.onInfo(Event.NewInfoNoOption(
		Event.SentMessage,
		"groupcasted websocketConnection message",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.GroupcastRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.GroupId:      groupId,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		}),
	))
	return nil
}
