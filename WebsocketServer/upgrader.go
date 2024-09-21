package WebsocketServer

import (
	"net"
	"net/http"

	"github.com/neutralusername/Systemge/Event"
)

func (server *WebsocketServer) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if event := server.onInfo(Event.New(
			Event.ExecutingHttpHandler,
			server.GetServerContext().Merge(Event.Context{
				"info":            "onConnect handler started",
				"httpHandlerType": "websocketUpgrade",
				"address":         httpRequest.RemoteAddr,
			}),
		)); event.IsError() {
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}

		ip, _, err := net.SplitHostPort(httpRequest.RemoteAddr)
		if err != nil {
			if event := server.onError(Event.New(
				Event.FailedToSplitHostPort,
				server.GetServerContext().Merge(Event.Context{
					"error":           "failed to split IP and port",
					"httpHandlerType": "websocketUpgrade",
					"address":         httpRequest.RemoteAddr,
				}),
			)); event.IsError() {
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			}
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}

		if server.ipRateLimiter != nil && !server.ipRateLimiter.RegisterConnectionAttempt(ip) {
			if event := server.onError(Event.New(
				Event.RateLimited,
				server.GetServerContext().Merge(Event.Context{
					"error":           "IP rate limit exceeded",
					"httpHandlerType": "websocketUpgrade",
					"rateLimiterType": "ip",
					"address":         httpRequest.RemoteAddr,
				}),
			)); event.IsError() {
				http.Error(responseWriter, "Rate limit exceeded", http.StatusTooManyRequests)
			}
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}

		websocketConnection, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			if event := server.onError(Event.New(
				Event.FailedToUpgradeToWebsocketConnection,
				server.GetServerContext().Merge(Event.Context{
					"error":           "failed to upgrade connection to websocket",
					"httpHandlerType": "websocketUpgrade",
					"address":         httpRequest.RemoteAddr,
				}),
			)); event.IsError() {
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			}
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}

		if event := server.onInfo(Event.New(
			Event.SendingToChannel,
			server.GetServerContext().Merge(Event.Context{
				"info":            "sending upgraded websocket connection to channel",
				"httpHandlerType": "websocketClient",
				"channelType":     "websocketConnection",
				"address":         websocketConnection.RemoteAddr().String(),
			}),
		)); event.IsError() {
			websocketConnection.Close()
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}
		server.connectionChannel <- websocketConnection
		server.acceptedWebsocketConnectionsCounter.Add(1)
		if event := server.onInfo(Event.New(
			Event.SentToChannel,
			server.GetServerContext().Merge(Event.Context{
				"info":            "sent upgraded websocket connection to channel",
				"httpHandlerType": "websocketUpgrade",
				"channelType":     "websocketConnection",
				"address":         httpRequest.RemoteAddr,
			}),
		)); event.IsError() {
			websocketConnection.Close()
		}

		server.onInfo(Event.New(
			Event.ExecutedHttpHandler,
			server.GetServerContext().Merge(Event.Context{
				"info":            "upgraded connection to websocket",
				"httpHandlerType": "websocketUpgrade",
				"address":         httpRequest.RemoteAddr,
			}),
		))
	}
}
