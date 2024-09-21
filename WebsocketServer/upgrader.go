package WebsocketServer

import (
	"net"
	"net/http"

	"github.com/neutralusername/Systemge/Event"
)

func (server *WebsocketServer) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if event := server.onInfo(Event.New(
			Event.HandlingHttpRequest,
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
			server.onError(Event.New(
				Event.FailedToSplitHostPort,
				server.GetServerContext().Merge(Event.Context{
					"error":           "failed to split IP and port",
					"httpHandlerType": "websocketUpgrade",
					"address":         httpRequest.RemoteAddr,
				}),
			))
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}

		if server.ipRateLimiter != nil && !server.ipRateLimiter.RegisterConnectionAttempt(ip) {
			server.onError(Event.New(
				Event.RateLimited,
				server.GetServerContext().Merge(Event.Context{
					"error":           "IP rate limit exceeded",
					"httpHandlerType": "websocketUpgrade",
					"rateLimiterType": "ip",
					"address":         httpRequest.RemoteAddr,
				}),
			))
			http.Error(responseWriter, "Rate limit exceeded", http.StatusTooManyRequests)
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}

		websocketConnection, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			server.onError(Event.New(
				Event.FailedToUpgradeToWebsocketConnection,
				server.GetServerContext().Merge(Event.Context{
					"error":           "failed to upgrade connection to websocket",
					"httpHandlerType": "websocketUpgrade",
					"address":         httpRequest.RemoteAddr,
				}),
			))
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}

		if event := server.sendWebsocketConnectionToChannel(websocketConnection); event.IsError() {
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}

		server.onInfo(Event.New(
			Event.HandledHttpRequest,
			server.GetServerContext().Merge(Event.Context{
				"info":            "upgraded connection to websocket",
				"httpHandlerType": "websocketUpgrade",
				"address":         httpRequest.RemoteAddr,
			}),
		))
	}
}
