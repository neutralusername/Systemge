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
			event := server.onError(Event.New(
				Event.FailedToSplitHostPort,
				server.GetServerContext().Merge(Event.Context{
					"error":   err.Error(),
					"address": httpRequest.RemoteAddr,
				}),
			))
			if event.IsError() {
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
				server.rejectedWebsocketConnectionsCounter.Add(1)
				return
			}
		}

		if server.ipRateLimiter != nil && !server.ipRateLimiter.RegisterConnectionAttempt(ip) {
			event := server.onError(Event.New(
				Event.RateLimited,
				server.GetServerContext().Merge(Event.Context{
					"error":   "IP rate limit exceeded",
					"type":    "ip",
					"address": httpRequest.RemoteAddr,
				}),
			))
			if event.IsError() {
				http.Error(responseWriter, "Rate limit exceeded", http.StatusTooManyRequests)
				server.rejectedWebsocketConnectionsCounter.Add(1)
				return
			}
		}

		websocketConnection, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			server.onError(Event.New(
				Event.FailedToUpgradeToWebsocketConnection,
				server.GetServerContext().Merge(Event.Context{
					"error":   err.Error(),
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
				server.GetServerContext().Merge(Event.Context{
					"error": "websocketServer stopped",
				}),
			))
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError) // idk if this will work after upgrade
			websocketConnection.Close()
			server.rejectedWebsocketConnectionsCounter.Add(1)
			server.waitGroup.Done()
			return
		default:
			if event := server.onInfo(Event.New(
				Event.SendingToChannel,
				server.GetServerContext().Merge(Event.Context{
					"info":    "sending upgraded websocket connection to channel",
					"type":    "websocketConnection",
					"address": websocketConnection.RemoteAddr().String(),
				}),
			)); event.IsError() {
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError) // idk if this will work after upgrade
				websocketConnection.Close()
				server.rejectedWebsocketConnectionsCounter.Add(1)
				server.waitGroup.Done()
			}
			server.connectionChannel <- websocketConnection
			server.onInfo(Event.New(
				Event.SentToChannel,
				server.GetServerContext().Merge(Event.Context{
					"info":    "sent upgraded websocket connection to channel",
					"type":    "websocketConnection",
					"address": websocketConnection.RemoteAddr().String(),
				}),
			))
		}
	}
}
