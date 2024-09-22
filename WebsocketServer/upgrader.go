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
				Event.FailedToSplitHostPort,
				err.Error(),
				server.GetServerContext().Merge(Event.Context{
					"address": httpRequest.RemoteAddr,
				}),
			))
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}

		if server.ipRateLimiter != nil && !server.ipRateLimiter.RegisterConnectionAttempt(ip) {
			event := server.onWarning(Event.NewWarning(
				Event.RateLimited,
				"websocket connection attempt rate limited",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Kind: "ip",
					"address":  httpRequest.RemoteAddr,
				}),
			))
			if !event.IsInfo() {
				http.Error(responseWriter, "Rate limit exceeded", http.StatusTooManyRequests)
				server.rejectedWebsocketConnectionsCounter.Add(1)
				return
			}
		}

		websocketConnection, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			server.onWarning(Event.NewWarningNoOption(
				Event.FailedToUpgradeToWebsocketConnection,
				err.Error(),
				server.GetServerContext().Merge(Event.Context{
					"address": httpRequest.RemoteAddr,
				}),
			))
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}

		server.waitGroup.Add(1)
		switch {
		case <-server.stopChannel:
			server.onInfo(Event.NewInfoNoOption(
				Event.ReceivedNilValueFromChannel,
				"rejecting connection because server is stopping",
				server.GetServerContext().Merge(Event.Context{
					Event.Kind: "stopChannel",
				}),
			))
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError) // idk if this will work after upgrade
			websocketConnection.Close()
			server.rejectedWebsocketConnectionsCounter.Add(1)
			server.waitGroup.Done()
			return
		default:
			if event := server.onInfo(Event.NewInfo(
				Event.SendingToChannel,
				"sending new websocket connection to channel",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Kind: "websocketConnection",
					"address":  websocketConnection.RemoteAddr().String(),
				}),
			)); !event.IsInfo() {
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError) // idk if this will work after upgrade
				websocketConnection.Close()
				server.rejectedWebsocketConnectionsCounter.Add(1)
				server.waitGroup.Done()
			}
			server.connectionChannel <- websocketConnection
			server.onInfo(Event.NewInfoNoOption(
				Event.SentToChannel,
				"sent upgraded websocket connection to channel",
				server.GetServerContext().Merge(Event.Context{
					Event.Kind: "websocketConnection",
					"address":  websocketConnection.RemoteAddr().String(),
				}),
			))
		}
	}
}
