package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

// Broadcast broadcasts a message to all connected clients.
// Blocking until all messages are sent.
func (server *WebsocketServer) Broadcast(message *Message.Message) error {
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	sessions := server.sessionManager.GetSessions("")
	targets := []string{}
	connections := []*WebsocketClient.WebsocketClient{}
	for _, session := range sessions {
		connection, ok := session.Get("websocketClient")
		if !ok {
			continue
		}
		websocketClient, ok := connection.(*WebsocketClient.WebsocketClient)
		if !ok {
			continue
		}
		connections = append(connections, websocketClient)
		id := websocketClient.GetName()
		if id != "" {
			targets = append(targets, id)
		}
	}

	targetsMarshalled := Helpers.JsonMarshal(targets)
	if server.eventHandler != nil {
		if event := server.onEvent(Event.New(
			Event.SendingBroadcast,
			Event.Context{
				Event.Targets:   targetsMarshalled,
				Event.Topic:     message.GetTopic(),
				Event.Payload:   message.GetPayload(),
				Event.SyncToken: message.GetSyncToken(),
			},
			Event.Continue,
			Event.Cancel,
		)); event.GetAction() == Event.Cancel {
			return errors.New("broadcast cancelled")
		}
	}

	for _, websocketClient := range connections {
		if websocketClient.GetName() == "" {
			if server.eventHandler != nil {
				event := server.onEvent(Event.New(
					Event.SessionNotAccepted,
					Event.Context{
						Event.Circumstance: Event.Broadcast,
						Event.SessionId:    websocketClient.GetName(),
						Event.Targets:      targetsMarshalled,
						Event.Topic:        message.GetTopic(),
						Event.Payload:      message.GetPayload(),
						Event.SyncToken:    message.GetSyncToken(),
					},
					Event.Skip,
					Event.Continue,
					Event.Cancel,
				))
				if event.GetAction() == Event.Cancel {
					return errors.New("broadcast cancelled")
				}
				if event.GetAction() == Event.Skip {
					continue
				}
			} else {
				continue
			}
		}
		waitGroup.AddTask(func() {
			websocketClient.Write(messageBytes)
		})
	}

	waitGroup.ExecuteTasksConcurrently()

	if server.eventHandler != nil {
		server.onEvent(Event.New(
			Event.SentBroadcast,
			Event.Context{
				Event.Circumstance: Event.Broadcast,
				Event.Targets:      targetsMarshalled,
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
			Event.Continue,
		))
	}
	return nil
}

// Multicast multicasts a message to multiple websocketConnections by id.
// Blocking until all messages are sent.
func (server *WebsocketServer) Multicast(ids []string, message *Message.Message) error {
	messageBytes := message.Serialize()
	waitGroup := Tools.NewTaskGroup()

	targetsMarshalled := Helpers.JsonMarshal(ids)
	if server.eventHandler != nil {
		if event := server.onEvent(Event.New(
			Event.SendingMulticast,
			Event.Context{
				Event.Circumstance: Event.Multicast,
				Event.Targets:      targetsMarshalled,
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
			Event.Continue,
			Event.Cancel,
		)); event.GetAction() == Event.Cancel {
			return errors.New("multicast cancelled")
		}
	}

	for _, id := range ids {
		session := server.sessionManager.GetSession(id)
		if session == nil {
			if server.eventHandler != nil {
				event := server.onEvent(Event.New(
					Event.SessionDoesNotExist,
					Event.Context{
						Event.Circumstance: Event.Multicast,
						Event.Target:       id,
						Event.Targets:      targetsMarshalled,
						Event.Topic:        message.GetTopic(),
						Event.Payload:      message.GetPayload(),
						Event.SyncToken:    message.GetSyncToken(),
					},
					Event.Skip,
					Event.Cancel,
				))
				if event.GetAction() == Event.Cancel {
					return errors.New("multicast cancelled")
				}
			}
			continue
		}

		websocketClient, ok := session.Get("connection")
		if !ok {
			// should never occur
			if server.eventHandler != nil {
				event := server.onEvent(Event.New(
					Event.SessionDoesNotExist,
					Event.Context{
						Event.Circumstance: Event.Multicast,
						Event.Target:       id,
						Event.Targets:      targetsMarshalled,
						Event.Topic:        message.GetTopic(),
						Event.Payload:      message.GetPayload(),
						Event.SyncToken:    message.GetSyncToken(),
					},
					Event.Skip,
					Event.Cancel,
				))
				if event.GetAction() == Event.Cancel {
					return errors.New("multicast cancelled")
				}
			}
			continue
		}

		if !session.IsAccepted() {
			if server.eventHandler != nil {
				event := server.onEvent(Event.New(
					Event.SessionNotAccepted,
					Event.Context{
						Event.Circumstance: Event.Multicast,
						Event.Target:       id,
						Event.Targets:      targetsMarshalled,
						Event.Topic:        message.GetTopic(),
						Event.Payload:      message.GetPayload(),
						Event.SyncToken:    message.GetSyncToken(),
					},
					Event.Continue,
					Event.Cancel,
				))
				if event.GetAction() == Event.Cancel {
					return errors.New("multicast cancelled")
				}
			}
		}

		waitGroup.AddTask(func() {
			websocketClient.(*WebsocketClient.WebsocketClient).Write(messageBytes)
		})
	}

	waitGroup.ExecuteTasksConcurrently()

	if server.eventHandler != nil {
		server.onEvent(Event.New(
			Event.SentMulticast,
			Event.Context{
				Event.Circumstance: Event.Multicast,
				Event.Targets:      targetsMarshalled,
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
			Event.Continue,
		))
	}
	return nil
}

func (server *WebsocketServer) Unicast(id string, message *Message.Message) error {
	return server.Multicast([]string{id}, message)
}
