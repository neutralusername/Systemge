package WebsocketServer

import (
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Event"
)

func (server *WebsocketServer) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if event := server.onInfo(Event.New(
			Event.HandlingHttpRequest,
			server.GetServerContext().Merge(Event.Context{
				"info":    "onConnect handler started",
				"type":    "websocketUpgrade",
				"address": httpRequest.RemoteAddr,
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
					"error":   err.Error(),
					"address": httpRequest.RemoteAddr,
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
					"error":   "IP rate limit exceeded",
					"type":    "ip",
					"address": httpRequest.RemoteAddr,
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
					"error":   err.Error(),
					"address": httpRequest.RemoteAddr,
				}),
			))
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}

		server.waitGroup.Add(1)
		if event := server.sendWebsocketConnectionToChannel(websocketConnection); event.IsError() {
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError) // idk if this will work after upgrade
			websocketConnection.Close()
			server.rejectedWebsocketConnectionsCounter.Add(1)
			server.waitGroup.Done()
			return
		}

		server.onInfo(Event.New(
			Event.HandledHttpRequest,
			server.GetServerContext().Merge(Event.Context{
				"info":    "upgraded connection to websocket",
				"type":    "websocketUpgrade",
				"address": httpRequest.RemoteAddr,
			}),
		))
	}
}

func (server *WebsocketServer) sendWebsocketConnectionToChannel(websocketConnection *websocket.Conn) *Event.Event {
	if event := server.onInfo(Event.New(
		Event.SendingToChannel,
		server.GetServerContext().Merge(Event.Context{
			"info":    "sending upgraded websocket connection to channel",
			"type":    "websocketConnection",
			"address": websocketConnection.RemoteAddr().String(),
		}),
	)); event.IsError() {
		return event
	}

	switch {
	case <-server.stopChannel:
		return server.onError(Event.New(
			Event.ServiceAlreadyStopped,
			server.GetServerContext().Merge(Event.Context{
				"error": "websocketServer stopped",
			}),
		))
	default:
		server.connectionChannel <- websocketConnection
		return server.onInfo(Event.New(
			Event.SentToChannel,
			server.GetServerContext().Merge(Event.Context{
				"info":    "sent upgraded websocket connection to channel",
				"type":    "websocketConnection",
				"address": websocketConnection.RemoteAddr().String(),
			}),
		))
	}
}
