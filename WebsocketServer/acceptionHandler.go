package WebsocketServer

import (
	"errors"
	"net"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

type AcceptionHandler[O any] func(*WebsocketServer[O], *WebsocketClient.WebsocketClient) (string, error)

func NewDefaultAcceptionHandler[O any]() AcceptionHandler[O] {
	return func(websocketServer *WebsocketServer[O], websocketClient *WebsocketClient.WebsocketClient) (string, error) {
		return "", nil
	}
}

func NewAccessControlAcceptionHandler[O any](blacklist *Tools.AccessControlList, whitelist *Tools.AccessControlList, ipRateLimiter *Tools.IpRateLimiter, handshakeHandler func(*WebsocketClient.WebsocketClient) (string, error)) AcceptionHandler[O] {
	return func(websocketServer *WebsocketServer[O], websocketClient *WebsocketClient.WebsocketClient) (string, error) {
		ip, _, err := net.SplitHostPort(websocketClient.GetAddress())
		if err != nil {
			if websocketServer.GetEventHandler() != nil {
				websocketServer.GetEventHandler().Handle(Event.New(
					Event.SplittingHostPortFailed,
					Event.Context{
						Event.Address: websocketClient.GetAddress(),
						Event.Error:   err.Error(),
					},
					Event.Skip,
				))
			}
			return "", err
		}

		if ipRateLimiter != nil && !ipRateLimiter.RegisterConnectionAttempt(ip) {
			if websocketServer.GetEventHandler() != nil {
				event := websocketServer.GetEventHandler().Handle(Event.New(
					Event.RateLimited,
					Event.Context{
						Event.Address:         websocketClient.GetAddress(),
						Event.RateLimiterType: Event.Ip,
					},
					Event.Skip,
					Event.Continue,
				))
				if event.GetAction() == Event.Skip {
					return "", errors.New("rate limited")
				}
			} else {
				return "", errors.New("rate limited")
			}
		}

		if blacklist != nil && blacklist.Contains(ip) {
			if websocketServer.GetEventHandler() != nil {
				event := websocketServer.GetEventHandler().Handle(Event.New(
					Event.Blacklisted,
					Event.Context{
						Event.Address: websocketClient.GetAddress(),
					},
					Event.Skip,
					Event.Continue,
				))
				if event.GetAction() == Event.Skip {
					return "", errors.New("blacklisted")
				}
			} else {
				return "", errors.New("blacklisted")
			}
		}

		if whitelist != nil && whitelist.ElementCount() > 0 && !whitelist.Contains(ip) {
			if websocketServer.GetEventHandler() != nil {
				event := websocketServer.GetEventHandler().Handle(Event.New(
					Event.NotWhitelisted,
					Event.Context{
						Event.Address: websocketClient.GetAddress(),
					},
					Event.Skip,
					Event.Continue,
				))
				if event.GetAction() == Event.Skip {
					return "", errors.New("not whitelisted")
				}
			} else {
				return "", errors.New("not whitelisted")
			}
		}

		identity := ""
		if handshakeHandler != nil {
			identity_, err := handshakeHandler(websocketClient)
			if err != nil {
				if websocketServer.GetEventHandler() != nil {
					event := websocketServer.GetEventHandler().Handle(Event.New(
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
						return "", err
					}
				} else {
					return "", err
				}
			}
			identity = identity_
		}

		return identity, nil
	}
}
