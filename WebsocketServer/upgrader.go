package WebsocketServer

import (
	"net"
	"net/http"

	"github.com/neutralusername/Systemge/Event"
)

func (server *WebsocketServer) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		ip, _, err := net.SplitHostPort(httpRequest.RemoteAddr)
		if err != nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.SplittingHostPortFailed,
				err.Error(),
				Event.Context{
					Event.Circumstance: Event.WebsocketUpgrade,
					Event.IdentityType: Event.HttpRequest,
					Event.Address:      httpRequest.RemoteAddr,
				}),
			)
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.websocketConnectionsFailed.Add(1)
			return
		}

		if server.ipRateLimiter != nil && !server.ipRateLimiter.RegisterConnectionAttempt(ip) {
			event := server.onEvent(Event.NewWarning(
				Event.RateLimited,
				"websocketConnection attempt ip rate limited",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance:    Event.WebsocketUpgrade,
					Event.RateLimiterType: Event.Ip,
					Event.IdentityType:    Event.HttpRequest,
					Event.Address:         httpRequest.RemoteAddr,
				}),
			)
			if !event.IsInfo() {
				http.Error(responseWriter, "Rate limit exceeded", http.StatusTooManyRequests)
				server.websocketConnectionsRejected.Add(1)
				return
			}
		}

		websocketConnection, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.WebsocketUpgradeFailed,
				err.Error(),
				Event.Context{
					Event.Circumstance: Event.WebsocketUpgrade,
					Event.IdentityType: Event.HttpRequest,
					Event.Address:      httpRequest.RemoteAddr,
				}),
			)
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.websocketConnectionsFailed.Add(1)
			return
		}

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
		server.connectionChannel <- websocketConnection
		server.onEvent(Event.NewInfoNoOption(
			Event.SentToChannel,
			"sent upgraded websocketConnection to channel",
			Event.Context{
				Event.Circumstance: Event.WebsocketUpgrade,
				Event.ChannelType:  Event.WebsocketConnection,
				Event.Address:      websocketConnection.RemoteAddr().String(),
			}),
		)
	}
}
