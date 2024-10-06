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
			server.createSession(websocketClient)
		} else {
			server.waitGroup.Add(1)
			go func(websocketClient *WebsocketClient.WebsocketClient) {
				server.createSession(websocketClient)
				server.waitGroup.Done()
			}(websocketClient)
		}
	}
}

func (server *WebsocketServer) createSession(websocketClient *WebsocketClient.WebsocketClient) {
	ip, _, err := net.SplitHostPort(websocketClient.GetAddress())
	if err != nil {
		if server.eventHandler != nil {
			event := server.onEvent(Event.New(
				Event.SplittingHostPortFailed,
				Event.Context{
					Event.Address: websocketClient.GetAddress(),
					Event.Error:   err.Error(),
				},
				Event.Skip,
				Event.Cancel,
			))
			if event.GetAction() == Event.Cancel {
				websocketClient.Close()
				break
			}
		}
		websocketClient.Close()
		continue
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
				Event.Cancel,
			))
			if event.GetAction() == Event.Cancel {
				websocketClient.Close()
				break
			}
			if event.GetAction() == Event.Skip {
				websocketClient.Close()
				continue
			}
		} else {
			websocketClient.Close()
			continue
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
				Event.Cancel,
			))
			if event.GetAction() == Event.Cancel {
				websocketClient.Close()
				break
			}
			if event.GetAction() == Event.Skip {
				websocketClient.Close()
				continue
			}
		} else {
			websocketClient.Close()
			continue
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
				Event.Cancel,
			))
			if event.GetAction() == Event.Cancel {
				websocketClient.Close()
				break
			}
			if event.GetAction() == Event.Skip {
				websocketClient.Close()
				continue
			}
		} else {
			websocketClient.Close()
			continue
		}
	}

	if server.eventHandler != nil {
		event := server.onEvent(Event.New(
			Event.CreatingSession,
			Event.Context{
				Event.Address: websocketClient.GetAddress(),
			},
			Event.Continue,
			Event.Skip,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			websocketClient.Close()
			break
		}
		if event.GetAction() == Event.Skip {
			websocketClient.Close()
			continue
		}
	}

	// add option to perform handshake to obtain identity

	session, err := server.sessionManager.CreateSession("", map[string]any{
		"websocketClient": websocketClient,
	})
	if err != nil {
		if server.eventHandler != nil {
			event := server.onEvent(Event.New(
				Event.CreateSessionFailed,
				Event.Context{
					Event.Address: websocketClient.GetAddress(),
				},
				Event.Skip,
				Event.Cancel,
			))
			if event.GetAction() == Event.Cancel {
				websocketClient.Close()
				break
			}
		}
		websocketClient.Close()
		continue
	}

	websocketClient.SetName(session.GetId()) // consider whether i can remove name from (Websocket-)Client
	if server.eventHandler != nil {
		event := server.onEvent(Event.New(
			Event.CreatedSession,
			Event.Context{
				Event.SessionId: session.GetId(),
				Event.Address:   websocketClient.GetAddress(),
			},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			websocketClient.Close()
			session.GetTimeout().Trigger()
			break
		}
	}

	server.waitGroup.Add(2)
	go server.websocketClientDisconnect(session, websocketClient)
	go server.receptionRoutine(session, websocketClient)
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
				Event.SessionId: session.GetId(),
				Event.Address:   websocketClient.(*WebsocketClient.WebsocketClient).GetAddress(),
			},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
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
				Event.Identity: session.GetId(),
				Event.Address:  websocketClient.GetAddress(),
			},
			Event.Continue,
		))
	}

	session.GetTimeout().Trigger()
	websocketClient.Close()
}
