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
			event := server.onWarning(Event.New(
				Event.FailedToSplitHostPort,
				err.Error(),
				Event.Warning,
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					"address": httpRequest.RemoteAddr,
				}),
			))
			if !event.IsInfo() {
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
				server.rejectedWebsocketConnectionsCounter.Add(1)
				return
			}
		}

		if server.ipRateLimiter != nil && !server.ipRateLimiter.RegisterConnectionAttempt(ip) {
			event := server.onWarning(Event.New(
				Event.RateLimited,
				"IP rate limit exceeded",
				Event.Warning,
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					"type":    "ip",
					"address": httpRequest.RemoteAddr,
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
			server.onWarning(Event.New(
				Event.FailedToUpgradeToWebsocketConnection,
				err.Error(),
				Event.Error,
				Event.NoOption,
				Event.NoOption,
				Event.NoOption,
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
			server.onError(Event.New(
				Event.ServiceAlreadyStopped,
				"websocketServer stopped",
				Event.Error,
				Event.NoOption,
				Event.NoOption,
				Event.NoOption,
				server.GetServerContext().Merge(Event.Context{}),
			))
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError) // idk if this will work after upgrade
			websocketConnection.Close()
			server.rejectedWebsocketConnectionsCounter.Add(1)
			server.waitGroup.Done()
			return
		default:
			if event := server.onInfo(Event.New(
				Event.SendingToChannel,
				"sending upgraded websocket connection to channel",
				Event.Info,
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					"type":    "websocketConnection",
					"address": websocketConnection.RemoteAddr().String(),
				}),
			)); !event.IsInfo() {
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError) // idk if this will work after upgrade
				websocketConnection.Close()
				server.rejectedWebsocketConnectionsCounter.Add(1)
				server.waitGroup.Done()
			}
			server.connectionChannel <- websocketConnection
			server.onInfo(Event.New(
				Event.SentToChannel,
				"sent upgraded websocket connection to channel",
				Event.Info,
				Event.NoOption,
				Event.NoOption,
				Event.NoOption,
				server.GetServerContext().Merge(Event.Context{
					"type":    "websocketConnection",
					"address": websocketConnection.RemoteAddr().String(),
				}),
			))
		}
	}
}
