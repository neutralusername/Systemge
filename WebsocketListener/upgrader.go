package WebsocketListener

import (
	"net"
	"net/http"

	"github.com/neutralusername/Systemge/Event"
)

func (server *WebsocketListener) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		ip, _, err := net.SplitHostPort(httpRequest.RemoteAddr)
		if err != nil {
			if server.eventHandler != nil {
				server.onEvent(Event.NewWarningNoOption(
					Event.SplittingHostPortFailed,
					err.Error(),
					Event.Context{
						Event.Circumstance: Event.WebsocketUpgrade,
						Event.Address:      httpRequest.RemoteAddr,
					}),
				)
			}
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.failed.Add(1)
			return
		}

		if server.ipRateLimiter != nil && !server.ipRateLimiter.RegisterConnectionAttempt(ip) {
			cancel := true
			if server.eventHandler != nil {
				if event := server.onEvent(Event.NewWarning(
					Event.RateLimited,
					"websocketConnection attempt ip rate limited",
					Event.Cancel,
					Event.Cancel,
					Event.Continue,
					Event.Context{
						Event.Circumstance:    Event.WebsocketUpgrade,
						Event.RateLimiterType: Event.Ip,
						Event.Address:         httpRequest.RemoteAddr,
					}),
				); event.IsInfo() {
					cancel = false
				}
			}
			if cancel {
				http.Error(responseWriter, "Rate limit exceeded", http.StatusTooManyRequests)
				server.rejected.Add(1)
				return
			}
		}

		websocketConnection, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			if server.eventHandler != nil {
				server.onEvent(Event.NewWarningNoOption(
					Event.WebsocketUpgradeFailed,
					err.Error(),
					Event.Context{
						Event.Circumstance: Event.WebsocketUpgrade,
						Event.Address:      httpRequest.RemoteAddr,
					}),
				)
			}
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.websocketConnectionsFailed.Add(1)
			return
		}
		if server.eventHandler != nil {
			if event := server.onEvent(Event.NewInfo(
				Event.SendingToChannel,
				"sending new websocketConnection to channel",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance: Event.WebsocketUpgrade,
					Event.ChannelType:  Event.WebsocketConnection,
					Event.Address:      websocketConnection.RemoteAddr().String(),
				}),
			); !event.IsInfo() {
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError) // idk if this will work after upgrade
				websocketConnection.Close()
				server.websocketConnectionsRejected.Add(1)
			}
		}
		server.connectionChannel <- websocketConnection
		if server.eventHandler != nil {
			server.onEvent(Event.NewInfoNoOption(
				Event.SentToChannel,
				"sent new websocketConnection to channel",
				Event.Context{
					Event.Circumstance: Event.WebsocketUpgrade,
					Event.ChannelType:  Event.WebsocketConnection,
					Event.Address:      websocketConnection.RemoteAddr().String(),
				}),
			)
		}
	}
}
