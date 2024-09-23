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
			server.onWarning(Event.NewWarningNoOption(
				Event.SplittingHostPortFailed,
				err.Error(),
				server.GetServerContext().Merge(Event.Context{
					Event.Circumstance:  Event.WebsocketUpgradeRoutine,
					Event.ClientAddress: httpRequest.RemoteAddr,
				}),
			))
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.websocketConnectionsFailed.Add(1)
			return
		}

		if server.ipRateLimiter != nil && !server.ipRateLimiter.RegisterConnectionAttempt(ip) {
			event := server.onWarning(Event.NewWarning(
				Event.RateLimited,
				"websocketConnection attempt ip rate limited",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Circumstance:    Event.WebsocketUpgradeRoutine,
					Event.RateLimiterType: Event.Ip,
					Event.ClientAddress:   httpRequest.RemoteAddr,
				}),
			))
			if !event.IsInfo() {
				http.Error(responseWriter, "Rate limit exceeded", http.StatusTooManyRequests)
				server.websocketConnectionsRejected.Add(1)
				return
			}
		}

		websocketConnection, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			server.onWarning(Event.NewWarningNoOption(
				Event.WebsocketUpgradeFailed,
				err.Error(),
				server.GetServerContext().Merge(Event.Context{
					Event.Circumstance:  Event.WebsocketUpgradeRoutine,
					Event.ClientAddress: httpRequest.RemoteAddr,
				}),
			))
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.websocketConnectionsFailed.Add(1)
			return
		}

		if event := server.onInfo(Event.NewInfo(
			Event.SendingToChannel,
			"sending new websocketConnection to channel",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:  Event.WebsocketUpgradeRoutine,
				Event.ChannelType:   Event.WebsocketConnection,
				Event.ClientAddress: websocketConnection.RemoteAddr().String(),
			}),
		)); !event.IsInfo() {
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError) // idk if this will work after upgrade
			websocketConnection.Close()
			server.websocketConnectionsRejected.Add(1)
		}
		server.connectionChannel <- websocketConnection
		server.onInfo(Event.NewInfoNoOption(
			Event.SentToChannel,
			"sent upgraded websocketConnection to channel",
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:  Event.WebsocketUpgradeRoutine,
				Event.ChannelType:   Event.WebsocketConnection,
				Event.ClientAddress: websocketConnection.RemoteAddr().String(),
			}),
		))
	}
}
