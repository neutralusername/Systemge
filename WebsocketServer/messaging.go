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
func (server *WebsocketServer) Unicast(sessionId string, message *Message.Message) error {
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	if event := server.onEvent(Event.NewInfo(
		Event.SendingMultiMessage,
		"unicasting websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Unicast,
			Event.Target:       sessionId,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	session := server.sessionManager.GetSession(sessionId)
	if session == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.SessionDoesNotExist,
			"session does not exist",
			Event.Context{
				Event.Circumstance: Event.Unicast,
				Event.Target:       sessionId,
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
		))
		return errors.New("session does not exist")
	}

	connection, ok := session.Get("connection")
	if !ok {
		// should never occur as of now
		server.onEvent(Event.NewWarningNoOption(
			Event.SessionDoesNotExist,
			"connection does not exist",
			Event.Context{
				Event.Circumstance: Event.Unicast,
				Event.Target:       sessionId,
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
		))
		return errors.New("connection does not exist")
	}
	websocketConnection, ok := connection.(*WebsocketConnection)

	if websocketConnection.GetId() == "" {
		event := server.onEvent(Event.NewWarning(
			Event.SessionNotAccepted,
			"websocketConnection is not accepted",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.Unicast,
				Event.Target:       sessionId,
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
		))
		if !event.IsInfo() {
			return event.GetError()
		}
	}

	waitGroup.AddTask(func() {
		server.write(websocketConnection, messageBytes, Event.Unicast)
	})

	waitGroup.ExecuteTasksConcurrently()

	server.onEvent(Event.NewInfoNoOption(
		Event.SentMultiMessage,
		"unicasted websocketConnection message",
		Event.Context{
			Event.Circumstance: Event.Unicast,
			Event.Target:       sessionId,
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

	targetsMarshalled := Helpers.JsonMarshal(ids)
	if event := server.onEvent(Event.NewInfo(
		Event.SendingMultiMessage,
		"multicasting websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Multicast,
			Event.Targets:      targetsMarshalled,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	for _, id := range ids {
		session := server.sessionManager.GetSession(id)
		if session == nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.SessionDoesNotExist,
				"session does not exist",
				Event.Context{
					Event.Circumstance: Event.Multicast,
					Event.Target:       id,
					Event.Targets:      targetsMarshalled,
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
			))
			continue
		}

		connection, ok := session.Get("connection")
		if !ok {
			// should never occur as of now
			server.onEvent(Event.NewWarningNoOption(
				Event.SessionDoesNotExist,
				"connection does not exist",
				Event.Context{
					Event.Circumstance: Event.Multicast,
					Event.Target:       id,
					Event.Targets:      targetsMarshalled,
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
			))
			continue
		}
		websocketConnection, ok := connection.(*WebsocketConnection)

		if websocketConnection.GetId() == "" {
			event := server.onEvent(Event.NewWarning(
				Event.SessionNotAccepted,
				"websocketConnection is not accepted",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance: Event.Multicast,
					Event.Target:       id,
					Event.Targets:      targetsMarshalled,
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
			))
			if !event.IsInfo() {
				continue
			}
		}

		waitGroup.AddTask(func() {
			server.write(websocketConnection, messageBytes, Event.Multicast)
		})
	}

	waitGroup.ExecuteTasksConcurrently()

	server.onEvent(Event.NewInfoNoOption(
		Event.SentMultiMessage,
		"multicasted websocketConnection message",
		Event.Context{
			Event.Circumstance: Event.Multicast,
			Event.Targets:      targetsMarshalled,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	))
	return nil
}
