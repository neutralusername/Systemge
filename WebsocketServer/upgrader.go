package WebsocketServer

import (
	"net"
	"net/http"

	"github.com/neutralusername/Systemge/Event"
)

func (server *WebsocketServer) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if server.httpServer == nil {
			server.onWarning(Event.New(
				Event.ServiceNotStarted,
				server.GetServerContext().Merge(Event.Context{
					"warning": "http server not started",
					"ip":      httpRequest.RemoteAddr,
				}),
			))
			return
		}
		ip, _, err := net.SplitHostPort(httpRequest.RemoteAddr)
		if err != nil {
			server.onWarning(Event.New(
				Event.FailedToSplitIPAndPort,
				server.GetServerContext().Merge(Event.Context{
					"waring": "failed to split IP and port",
					"ip":     httpRequest.RemoteAddr,
				}),
			))
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			return
		}
		if server.ipRateLimiter != nil && !server.ipRateLimiter.RegisterConnectionAttempt(ip) {
			server.onWarning(Event.New(
				Event.IPRateLimitExceeded,
				server.GetServerContext().Merge(Event.Context{
					"warning": "IP rate limit exceeded",
					"ip":      ip,
				}),
			))
			http.Error(responseWriter, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		websocketConnection, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			server.onWarning(Event.New(
				Event.FailedToUpgradeToWebsocket,
				server.GetServerContext().Merge(Event.Context{
					"warning": "failed to upgrade connection to websocket",
					"ip":      ip,
				}),
			))
			return
		}
		server.onInfo(Event.New(
			Event.UpgradedToWebsocket,
			server.GetServerContext().Merge(Event.Context{
				"info": "upgraded connection to websocket",
				"ip":   ip,
			}),
		))
		server.connectionChannel <- websocketConnection
	}
}
