package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (server *WebsocketServer) acceptionRoutine() {
	defer func() {
		if server.eventHandler != nil {
			server.eventHandler.Handle(Event.New(
				Event.AcceptionRoutineEnds,
				Event.Context{},
				Event.Continue,
			))
		}
		server.waitGroup.Done()
	}()

	if server.eventHandler != nil {
		event := server.eventHandler.Handle(Event.New(
			Event.AcceptionRoutineBegins,
			Event.Context{},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			return
		}
	}

	handleAcceptionWrapper := func(websocketClient *WebsocketClient.WebsocketClient) {
		if identity, err := server.acceptionHandler(websocketClient); err == nil {
			session := server.createSession(identity, websocketClient)
			if session == nil {
				server.ClientsRejected.Add(1)
				websocketClient.Close()
				return
			}
			server.ClientsAccepted.Add(1)
			server.waitGroup.Add(2)
			go server.websocketClientDisconnect(session, websocketClient)
			go server.receptionRoutine(session, websocketClient)
		} else {
			server.ClientsRejected.Add(1)
			websocketClient.Close()
		}
	}

	for {
		websocketClient, err := server.websocketListener.Accept(server.config.WebsocketClientConfig, server.config.AcceptTimeoutMs)
		if err != nil {
			websocketClient.Close()
			if server.eventHandler != nil {
				event := server.eventHandler.Handle(Event.New(
					Event.AcceptClientFailed,
					Event.Context{
						Event.Address: websocketClient.GetAddress(),
					},
					Event.Skip,
					Event.Cancel,
				))
				if event.GetAction() == Event.Cancel {
					break
				}
			}
			continue
		}

		if server.config.HandleClientsSequentially {
			handleAcceptionWrapper(websocketClient)
		} else {
			server.waitGroup.Add(1)
			go func(websocketClient *WebsocketClient.WebsocketClient) {
				handleAcceptionWrapper(websocketClient)
				server.waitGroup.Done()
			}(websocketClient)
		}
	}
}

func (server *WebsocketServer) createSession(identity string, websocketClient *WebsocketClient.WebsocketClient) *Tools.Session {
	for {
		if server.eventHandler != nil {
			event := server.eventHandler.Handle(Event.New(
				Event.CreatingSession,
				Event.Context{
					Event.Address: websocketClient.GetAddress(),
				},
				Event.Continue,
				Event.Skip,
			))
			if event.GetAction() == Event.Skip {
				return nil
			}
		}

		session, err := server.sessionManager.CreateSession(identity, map[string]any{
			"websocketClient": websocketClient,
		})
		if err != nil {
			if server.eventHandler != nil {
				event := server.eventHandler.Handle(Event.New(
					Event.CreateSessionFailed,
					Event.Context{
						Event.Address:  websocketClient.GetAddress(),
						Event.Identity: identity,
					},
					Event.Skip,
					Event.Retry,
				))
				if event.GetAction() == Event.Retry {
					continue
				}
			}
			return nil
		}

		if server.eventHandler != nil {
			event := server.eventHandler.Handle(Event.New(
				Event.CreatedSession,
				Event.Context{
					Event.Address:   websocketClient.GetAddress(),
					Event.SessionId: session.GetId(),
					Event.Identity:  session.GetIdentity(),
				},
				Event.Continue,
				Event.Skip,
				Event.Retry,
			))
			if event.GetAction() == Event.Skip {
				session.GetTimeout().Trigger()
				return nil
			}
			if event.GetAction() == Event.Retry {
				session.GetTimeout().Trigger()
				continue
			}
		}
		return session
	}
}

func (server *WebsocketServer) onCreateSession(session *Tools.Session) error {
	websocketClient, ok := session.Get("websocketClient")
	if !ok {
		return errors.New("websocketClient not found")
	}

	if server.eventHandler != nil {
		event := server.eventHandler.Handle(Event.New(
			Event.OnCreateSession,
			Event.Context{
				Event.Address:   websocketClient.(*WebsocketClient.WebsocketClient).GetAddress(),
				Event.Identity:  session.GetIdentity(),
				Event.SessionId: session.GetId(),
			},
			Event.Continue,
			Event.Skip,
		))
		if event.GetAction() == Event.Skip {
			return errors.New("session rejected")
		}
	}

	return nil
}

func (server *WebsocketServer) websocketClientDisconnect(session *Tools.Session, websocketClient *WebsocketClient.WebsocketClient) {
	defer server.waitGroup.Done()

	select {
	case <-websocketClient.GetCloseChannel():
	case <-session.GetTimeout().GetTriggeredChannel():
	case <-server.stopChannel:
	}

	if server.eventHandler != nil {
		server.eventHandler.Handle(Event.New(
			Event.OnDisconnect,
			Event.Context{
				Event.Address:   websocketClient.GetAddress(),
				Event.Identity:  session.GetId(),
				Event.SessionId: session.GetIdentity(),
			},
			Event.Continue,
		))
	}

	session.GetTimeout().Trigger()
	websocketClient.Close()
}

/*
	identity := ""
	if server.handshakeHandler != nil {
		identity_, err := server.handshakeHandler(websocketClient)
		if err != nil {
			if server.eventHandler != nil {
				event := server.onEvent(Event.New(
					Event.HandshakeFailed,
					Event.Context{
						Event.Address:  websocketClient.GetAddress(),
						Event.Identity: identity,
						Event.Error:    err.Error(),
					},
					Event.Skip,
					Event.Continue,
				))
				if event.GetAction() == Event.Skip {
					return err
				}
			} else {
				return err
			}
		}
		identity = identity_
	}
*/
