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
		if server.httpServer == nil {
			if event := server.onError(Event.New(
				Event.FailedToExecuteHttpHandler,
				server.GetServerContext().Merge(Event.Context{
					"error":           "accepted connection on websocketServer but httpServer is nil",
					"httpHandlerType": "websocketUpgrade",
					"address":         httpRequest.RemoteAddr,
				}),
			)); event.IsError() {
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			}
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}
		ip, _, err := net.SplitHostPort(httpRequest.RemoteAddr)
		if err != nil {
			if event := server.onError(Event.New(
				Event.FailedToExecuteHttpHandler,
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
				Event.FailedToExecuteHttpHandler,
				server.GetServerContext().Merge(Event.Context{
					"error":           "IP rate limit exceeded",
					"httpHandlerType": "websocketUpgrade",
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
				Event.FailedToExecuteHttpHandler,
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
			Event.ExecutedHttpHandler,
			server.GetServerContext().Merge(Event.Context{
				"info":            "upgraded connection to websocket",
				"httpHandlerType": "websocketUpgrade",
				"address":         httpRequest.RemoteAddr,
			}),
		)); event.IsError() {
			websocketConnection.Close()
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}
		server.acceptedWebsocketConnectionsCounter.Add(1)
		server.connectionChannel <- websocketConnection
	}
}
