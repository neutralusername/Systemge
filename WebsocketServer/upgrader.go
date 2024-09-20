package WebsocketServer

import (
	"net"
	"net/http"

	"github.com/neutralusername/Systemge/Event"
)

func (server *WebsocketServer) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		server.onInfo(Event.New(
			Event.OnConnectHandlerStarted,
			server.GetServerContext().Merge(Event.Context{
				"info":                 "onConnect handler started",
				"onConnectHandlerType": "httpWebsocketUpgrade",
				"ip":                   httpRequest.RemoteAddr,
			}),
		))
		defer server.onInfo(Event.New(
			Event.OnConnectHandlerFinished,
			server.GetServerContext().Merge(Event.Context{
				"info":                 "onConnect handler finished",
				"onConnectHandlerType": "httpWebsocketUpgrade",
				"ip":                   httpRequest.RemoteAddr,
			}),
		))
		if server.httpServer == nil {
			server.onWarning(Event.New(
				Event.ServiceNotStarted,
				server.GetServerContext().Merge(Event.Context{
					"warning":              "http server not started",
					"onConnectHandlerType": "httpWebsocketUpgrade",
					"ip":                   httpRequest.RemoteAddr,
				}),
			))
			return
		}
		ip, _, err := net.SplitHostPort(httpRequest.RemoteAddr)
		if err != nil {
			server.onWarning(Event.New(
				Event.FailedToSplitIPAndPort,
				server.GetServerContext().Merge(Event.Context{
					"waring":               "failed to split IP and port",
					"onConnectHandlerType": "httpWebsocketUpgrade",
					"ip":                   httpRequest.RemoteAddr,
				}),
			))
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			return
		}
		if server.ipRateLimiter != nil && !server.ipRateLimiter.RegisterConnectionAttempt(ip) {
			server.onWarning(Event.New(
				Event.IPRateLimitExceeded,
				server.GetServerContext().Merge(Event.Context{
					"warning":              "IP rate limit exceeded",
					"onConnectHandlerType": "httpWebsocketUpgrade",
					"ip":                   ip,
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
					"warning":              "failed to upgrade connection to websocket",
					"onConnectHandlerType": "httpWebsocketUpgrade",
					"ip":                   ip,
				}),
			))
			return
		}
		server.onInfo(Event.New(
			Event.UpgradedToWebsocket,
			server.GetServerContext().Merge(Event.Context{
				"info":                 "upgraded connection to websocket",
				"onConnectHandlerType": "httpWebsocketUpgrade",
				"ip":                   ip,
			}),
		))
		server.connectionChannel <- websocketConnection
	}
}
