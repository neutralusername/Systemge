package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (server *WebsocketServer) Message(messageBytes []byte, ids ...string) error {
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
