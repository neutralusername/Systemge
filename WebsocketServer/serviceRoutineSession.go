package WebsocketServer

import (
	"errors"
	"net"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (server *WebsocketServer) sessionRoutine() {
	defer func() {
		if server.eventHandler != nil {
			server.onEvent(Event.New(
				Event.SessionRoutineEnds,
				Event.Context{},
				Event.Continue,
				Event.Cancel,
			))
		}
		server.waitGroup.Done()
	}()

	if server.eventHandler != nil {
		event := server.onEvent(Event.New(
			Event.SessionRoutineBegins,
			Event.Context{},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			return
		}
	}

	for {
		websocketClient, err := server.websocketListener.Accept(server.config.WebsocketClientConfig, server.eventHandler)
		if err != nil {
			websocketClient.Close()
			if server.eventHandler != nil {
				event := server.onEvent(Event.New(
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

		if server.config.AcceptClientsSequentially {
			server.handleAccept(websocketClient)
		} else {
			server.waitGroup.Add(1)
			go func(websocketClient *WebsocketClient.WebsocketClient) {
				server.handleAccept(websocketClient)
				server.waitGroup.Done()
			}(websocketClient)
		}
	}
}

func (server *WebsocketServer) handleAccept(websocketClient *WebsocketClient.WebsocketClient) {
	ip, _, err := net.SplitHostPort(websocketClient.GetAddress())
	if err != nil {
		if server.eventHandler != nil {
			server.onEvent(Event.New(
				Event.SplittingHostPortFailed,
				Event.Context{
					Event.Address: websocketClient.GetAddress(),
					Event.Error:   err.Error(),
				},
				Event.Skip,
			))
		}
		websocketClient.Close()
		return
	}

	if server.ipRateLimiter != nil && !server.ipRateLimiter.RegisterConnectionAttempt(ip) {
		if server.eventHandler != nil {
			event := server.onEvent(Event.New(
				Event.RateLimited,
				Event.Context{
					Event.Address:         websocketClient.GetAddress(),
					Event.RateLimiterType: Event.Ip,
				},
				Event.Skip,
				Event.Continue,
			))
			if event.GetAction() == Event.Skip {
				websocketClient.Close()
				return
			}
		} else {
			websocketClient.Close()
			return
		}
	}

	if server.blacklist != nil && server.blacklist.Contains(ip) {
		if server.eventHandler != nil {
			event := server.onEvent(Event.New(
				Event.Blacklisted,
				Event.Context{
					Event.Address: websocketClient.GetAddress(),
				},
				Event.Skip,
				Event.Continue,
			))
			if event.GetAction() == Event.Skip {
				websocketClient.Close()
				return
			}
		} else {
			websocketClient.Close()
			return
		}
	}

	if server.whitelist != nil && server.whitelist.ElementCount() > 0 && !server.whitelist.Contains(ip) {
		if server.eventHandler != nil {
			event := server.onEvent(Event.New(
				Event.NotWhitelisted,
				Event.Context{
					Event.Address: websocketClient.GetAddress(),
				},
				Event.Skip,
				Event.Continue,
			))
			if event.GetAction() == Event.Skip {
				websocketClient.Close()
				return
			}
		} else {
			websocketClient.Close()
			return
		}
	}
	identity := ""
	if server.handshakeHandler != nil {
		id, err := server.handshakeHandler(websocketClient)
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
					websocketClient.Close()
					return
				}
			} else {
				websocketClient.Close()
				return
			}
		}
		identity = id
	}

	session := server.createSession(identity, websocketClient)
	if session == nil {
		websocketClient.Close()
		return
	}

	server.waitGroup.Add(2)
	go server.websocketClientDisconnect(session, websocketClient)
	go server.receptionRoutine(session, websocketClient)
}

func (server *WebsocketServer) createSession(identity string, websocketClient *WebsocketClient.WebsocketClient) *Tools.Session {
	for {
		if server.eventHandler != nil {
			event := server.onEvent(Event.New(
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
				event := server.onEvent(Event.New(
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
			event := server.onEvent(Event.New(
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
		event := server.onEvent(Event.New(
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
		server.onEvent(Event.New(
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
