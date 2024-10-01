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

	sessions := server.sessionManager.GetSessions("")
	targets := []string{}
	connections := []*WebsocketConnection{}
	for _, session := range sessions {
		connection, ok := session.Get("connection")
		if !ok {
			continue
		}
		websocketConnection, ok := connection.(*WebsocketConnection)
		if !ok {
			continue
		}
		connections = append(connections, websocketConnection)
		id := websocketConnection.GetId()
		if id != "" {
			targets = append(targets, id)
		}
	}

	targetsMarshalled := Helpers.JsonMarshal(targets)
	if event := server.onEvent(Event.NewInfo(
		Event.SendingMultiMessage,
		"broadcasting websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Broadcast,
			Event.Targets:      targetsMarshalled,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	for _, websocketConnection := range connections {
		if websocketConnection.GetId() == "" {
			event := server.onEvent(Event.NewWarning(
				Event.SessionNotAccepted,
				"websocketConnection is not accepted",
				Event.Cancel,
				Event.Skip,
				Event.Continue,
				Event.Context{
					Event.Circumstance: Event.Broadcast,
					Event.SessionId:    websocketConnection.GetId(),
					Event.Targets:      targetsMarshalled,
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
			))
			if event.IsError() {
				return event.GetError()
			}
			if event.IsWarning() {
				continue
			}
		}
		waitGroup.AddTask(func() {
			server.write(websocketConnection, messageBytes, Event.Broadcast)
		})
	}

	waitGroup.ExecuteTasksConcurrently()

	server.onEvent(Event.NewInfoNoOption(
		Event.SentMultiMessage,
		"broadcasted websocketConnection message",
		Event.Context{
			Event.Circumstance: Event.Broadcast,
			Event.Targets:      targetsMarshalled,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	))
	return nil
}

// Unicast unicasts a message to a specific websocketConnection by id.
// Blocking until the message is sent.
func (server *WebsocketServer) Unicast(websocketId string, message *Message.Message) error {
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.websocketConnectionMutex.RLock()
	if event := server.onEvent(Event.NewInfo(
		Event.SendingMultiMessage,
		"unicasting websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Unicast,
			Event.IdentityType: Event.WebsocketConnection,
			Event.Target:       websocketId,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	)); !event.IsInfo() {
		server.websocketConnectionMutex.RUnlock()
		return event.GetError()
	}

	websocketConnection, exists := server.websocketConnections[websocketId]
	if !exists {
		server.websocketConnectionMutex.RUnlock()
		server.onEvent(Event.NewWarningNoOption(
			Event.ClientDoesNotExist,
			"websocketConnection does not exist",
			Event.Context{
				Event.Circumstance: Event.Unicast,
				Event.IdentityType: Event.WebsocketConnection,
				Event.ClientId:     websocketId,
				Event.Target:       websocketId,
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
		))
		return errors.New("websocketConnection does not exist")
	}
	if !websocketConnection.isAccepted {
		event := server.onEvent(Event.NewWarning(
			Event.ClientNotAccepted,
			"websocketConnection is not accepted",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.Unicast,
				Event.IdentityType: Event.WebsocketConnection,
				Event.ClientId:     websocketId,
				Event.Target:       websocketId,
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
		))
		if !event.IsInfo() {
			server.websocketConnectionMutex.RUnlock()
			return event.GetError()
		}
	}
	waitGroup.AddTask(func() {
		server.write(websocketConnection, messageBytes, Event.Unicast)
	})
	server.websocketConnectionMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()

	server.onEvent(Event.NewInfoNoOption(
		Event.SentMultiMessage,
		"unicasted websocketConnection message",
		Event.Context{
			Event.Circumstance: Event.Unicast,
			Event.IdentityType: Event.WebsocketConnection,
			Event.Target:       websocketId,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
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
	if event := server.onEvent(Event.NewInfo(
		Event.SendingMultiMessage,
		"multicasting websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Multicast,
			Event.IdentityType: Event.WebsocketConnection,
			Event.Targets:      targetClientIds,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	)); !event.IsInfo() {
		server.websocketConnectionMutex.RUnlock()
		return event.GetError()
	}

	for _, id := range ids {
		websocketConnection, exists := server.websocketConnections[id]
		if !exists {
			event := server.onEvent(Event.NewWarning(
				Event.ClientDoesNotExist,
				"websocketConnection does not exist",
				Event.Cancel,
				Event.Cancel,
				Event.Skip,
				Event.Context{
					Event.Circumstance: Event.Multicast,
					Event.IdentityType: Event.WebsocketConnection,
					Event.ClientId:     id,
					Event.Targets:      targetClientIds,
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
			))
			if event.IsError() {
				server.websocketConnectionMutex.RUnlock()
				return event.GetError()
			} else {
				continue
			}
		}
		if !websocketConnection.isAccepted {
			event := server.onEvent(Event.NewWarning(
				Event.ClientNotAccepted,
				"websocketConnection is not accepted",
				Event.Cancel,
				Event.Skip,
				Event.Continue,
				Event.Context{
					Event.Circumstance: Event.Multicast,
					Event.IdentityType: Event.WebsocketConnection,
					Event.ClientId:     id,
					Event.Targets:      targetClientIds,
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
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
			server.write(websocketConnection, messageBytes, Event.Multicast)
		})
	}
	server.websocketConnectionMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()

	server.onEvent(Event.NewInfoNoOption(
		Event.SentMultiMessage,
		"multicasted websocketConnection message",
		Event.Context{
			Event.Circumstance: Event.Multicast,
			Event.IdentityType: Event.WebsocketConnection,
			Event.Targets:      targetClientIds,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	))
	return nil
}

// Groupcast groupcasts a message to all websocketConnections in a group.
// Blocking until all messages are sent.
func (server *WebsocketServer) Groupcast(groupId string, message *Message.Message) error {
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	server.websocketConnectionMutex.RLock()
	if event := server.onEvent(Event.NewInfo(
		Event.SendingMultiMessage,
		"groupcasting websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Groupcast,
			Event.IdentityType: Event.WebsocketConnection,
			Event.GroupId:      groupId,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	)); !event.IsInfo() {
		server.websocketConnectionMutex.RUnlock()
		return event.GetError()
	}

	group, ok := server.groupsWebsocketConnections[groupId]
	if !ok {
		server.websocketConnectionMutex.RUnlock()
		server.onEvent(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			Event.Context{
				Event.Circumstance: Event.Groupcast,
				Event.IdentityType: Event.WebsocketConnection,
				Event.GroupId:      groupId,
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
		))
		return errors.New("group does not exist")
	}
	for _, websocketConnection := range group {
		if !websocketConnection.isAccepted {
			event := server.onEvent(Event.NewWarning(
				Event.ClientNotAccepted,
				"websocketConnection is not accepted",
				Event.Cancel,
				Event.Skip,
				Event.Continue,
				Event.Context{
					Event.Circumstance: Event.Groupcast,
					Event.IdentityType: Event.WebsocketConnection,
					Event.ClientId:     websocketConnection.GetId(),
					Event.GroupId:      groupId,
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
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
			server.write(websocketConnection, messageBytes, Event.Groupcast)
		})
	}
	server.websocketConnectionMutex.RUnlock()

	waitGroup.ExecuteTasksConcurrently()

	server.onEvent(Event.NewInfoNoOption(
		Event.SentMultiMessage,
		"groupcasted websocketConnection message",
		Event.Context{
			Event.Circumstance: Event.Groupcast,
			Event.IdentityType: Event.WebsocketConnection,
			Event.GroupId:      groupId,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	))
	return nil
}
