package WebsocketServer

import (
	"net"
	"net/http"

	"github.com/neutralusername/Systemge/Event"
)

func (server *WebsocketServer) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if event := server.onInfo(Event.New(
			Event.OnConnectHandlerStarted,
			server.GetServerContext().Merge(Event.Context{
				"info":                 "onConnect handler started",
				"onConnectHandlerType": "httpWebsocketUpgrade",
				"address":              httpRequest.RemoteAddr,
			}),
		)); event.IsError() {
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			return
		}
		if server.httpServer == nil {
			if event := server.onError(Event.New(
				Event.ServiceNotStarted,
				server.GetServerContext().Merge(Event.Context{
					"error":                "accepted connection on websocketServer but httpServer is nil",
					"onConnectHandlerType": "httpWebsocketUpgrade",
					"address":              httpRequest.RemoteAddr,
				}),
			)); event.IsError() {
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			}
			return
		}
		ip, _, err := net.SplitHostPort(httpRequest.RemoteAddr)
		if err != nil {
			if event := server.onError(Event.New(
				Event.FailedToSplitIPAndPort,
				server.GetServerContext().Merge(Event.Context{
					"error":                "failed to split IP and port",
					"onConnectHandlerType": "httpWebsocketUpgrade",
					"address":              httpRequest.RemoteAddr,
				}),
			)); event.IsError() {
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			}
			return
		}
		if server.ipRateLimiter != nil && !server.ipRateLimiter.RegisterConnectionAttempt(ip) {
			if event := server.onError(Event.New(
				Event.IPRateLimitExceeded,
				server.GetServerContext().Merge(Event.Context{
					"error":                "IP rate limit exceeded",
					"onConnectHandlerType": "httpWebsocketUpgrade",
					"address":              httpRequest.RemoteAddr,
				}),
			)); event.IsError() {
				http.Error(responseWriter, "Rate limit exceeded", http.StatusTooManyRequests)
			}
			return
		}
		websocketConnection, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			if event := server.onError(Event.New(
				Event.FailedToUpgradeToWebsocket,
				server.GetServerContext().Merge(Event.Context{
					"error":                "failed to upgrade connection to websocket",
					"onConnectHandlerType": "httpWebsocketUpgrade",
					"address":              httpRequest.RemoteAddr,
				}),
			)); event.IsError() {
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			}
			return
		}
		if event := server.onInfo(Event.New(
			Event.OnConnectHandlerFinished,
			server.GetServerContext().Merge(Event.Context{
				"info":                 "onConnect handler finished",
				"onConnectHandlerType": "httpWebsocketUpgrade",
				"address":              httpRequest.RemoteAddr,
			}),
		)); event.IsError() {
			websocketConnection.Close()
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			return
		}
		server.connectionChannel <- websocketConnection
	}
}
