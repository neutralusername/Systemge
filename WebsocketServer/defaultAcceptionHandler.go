package WebsocketServer

import (
	"errors"
	"net"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (server *WebsocketServer) GetDefaultAcceptionHandler() func(*WebsocketClient.WebsocketClient) (string, error) {
	return func(websocketClient *WebsocketClient.WebsocketClient) (string, error) {
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
			return "", err
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
					return "", errors.New("rate limited")
				}
			} else {
				return "", errors.New("rate limited")
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
					return "", errors.New("blacklisted")
				}
			} else {
				return "", errors.New("blacklisted")
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
					return "", errors.New("not whitelisted")
				}
			} else {
				return "", errors.New("not whitelisted")
			}
		}

		return "", nil
	}
}
