package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (server *WebsocketServer) Message(messageBytes []byte, ids ...string) error {
	if len(ids) == 0 {
		return server.broadcast(messageBytes)
	}
	waitGroup := Tools.NewTaskGroup()

	targetsMarshalled := Helpers.JsonMarshal(ids)
	if server.eventHandler != nil {
		if event := server.eventHandler.Handle(Event.New(
			Event.SendingMulticast,
			Event.Context{
				Event.Circumstance: Event.Multicast,
				Event.Targets:      targetsMarshalled,
				Event.Bytes:        string(messageBytes),
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
				event := server.eventHandler.Handle(Event.New(
					Event.SessionDoesNotExist,
					Event.Context{
						Event.Circumstance: Event.Multicast,
						Event.Target:       id,
						Event.Targets:      targetsMarshalled,
						Event.Bytes:        string(messageBytes),
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

		websocketClient, ok := session.Get("websocketClient")
		if !ok {
			// should never occur
			if server.eventHandler != nil {
				event := server.eventHandler.Handle(Event.New(
					Event.SessionDoesNotExist,
					Event.Context{
						Event.Circumstance: Event.Multicast,
						Event.Target:       id,
						Event.Targets:      targetsMarshalled,
						Event.Bytes:        string(messageBytes),
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
				event := server.eventHandler.Handle(Event.New(
					Event.SessionNotAccepted,
					Event.Context{
						Event.Circumstance: Event.Multicast,
						Event.Target:       id,
						Event.Targets:      targetsMarshalled,
						Event.Bytes:        string(messageBytes),
					},
					Event.Continue,
					Event.Skip,
					Event.Cancel,
				))
				if event.GetAction() == Event.Skip {
					continue
				}
				if event.GetAction() == Event.Cancel {
					return errors.New("multicast cancelled")
				}
			}
		}

		waitGroup.AddTask(func() {
			websocketClient.(*WebsocketClient.WebsocketClient).Write(messageBytes, server.config.WriteTimeoutMs)
		})
	}

	waitGroup.ExecuteTasksConcurrently()

	if server.eventHandler != nil {
		server.eventHandler.Handle(Event.New(
			Event.SentMulticast,
			Event.Context{
				Event.Circumstance: Event.Multicast,
				Event.Targets:      targetsMarshalled,
				Event.Bytes:        string(messageBytes),
			},
			Event.Continue,
		))
	}
	return nil
}

func (server *WebsocketServer) broadcast(messageBytes []byte) error {
	waitGroup := Tools.NewTaskGroup()

	sessions := server.sessionManager.GetSessions()

	if server.eventHandler != nil {
		if event := server.eventHandler.Handle(Event.New(
			Event.SendingBroadcast,
			Event.Context{
				Event.Bytes: string(messageBytes),
			},
			Event.Continue,
			Event.Cancel,
		)); event.GetAction() == Event.Cancel {
			return errors.New("broadcast cancelled")
		}
	}

	for _, session := range sessions {
		if !session.IsAccepted() {
			if server.eventHandler != nil {
				event := server.eventHandler.Handle(Event.New(
					Event.SessionNotAccepted,
					Event.Context{
						Event.Circumstance: Event.Broadcast,
						Event.Target:       session.GetId(),
						Event.Bytes:        string(messageBytes),
					},
					Event.Skip,
					Event.Continue,
					Event.Cancel,
				))
				if event.GetAction() == Event.Skip {
					continue
				}
				if event.GetAction() == Event.Cancel {
					return errors.New("broadcast cancelled")
				}
			} else {
				continue
			}
		}

		websocketClient, ok := session.Get("websocketClient")
		if !ok {
			// should never occur
			if server.eventHandler != nil {
				event := server.eventHandler.Handle(Event.New(
					Event.SessionDoesNotExist,
					Event.Context{
						Event.Circumstance: Event.Broadcast,
						Event.Target:       session.GetId(),
						Event.Bytes:        string(messageBytes),
					},
					Event.Skip,
					Event.Cancel,
				))
				if event.GetAction() == Event.Cancel {
					return errors.New("broadcast cancelled")
				}
			}
			continue
		}

		waitGroup.AddTask(func() {
			websocketClient.(*WebsocketClient.WebsocketClient).Write(messageBytes, server.config.WriteTimeoutMs)
		})
	}

	waitGroup.ExecuteTasksConcurrently()

	if server.eventHandler != nil {
		server.eventHandler.Handle(Event.New(
			Event.SentBroadcast,
			Event.Context{
				Event.Circumstance: Event.Broadcast,
				Event.Bytes:        string(messageBytes),
			},
			Event.Continue,
		))
	}
	return nil
}
