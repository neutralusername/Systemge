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
				Event.SplittingHostPortFailed,
				err.Error(),
				server.GetServerContext().Merge(Event.Context{
					Event.Address: httpRequest.RemoteAddr,
				}),
			))
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.rejectedWebsocketConnectionsCounter.Add(1)
			return
		}

		if server.ipRateLimiter != nil && !server.ipRateLimiter.RegisterConnectionAttempt(ip) {
			event := server.onWarning(Event.NewWarning(
				Event.RateLimited,
				"websocketConnection attempt ip rate limited",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:    Event.Ip,
					Event.Address: httpRequest.RemoteAddr,
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
				Event.WebsocketUpgradeFailed,
				err.Error(),
				server.GetServerContext().Merge(Event.Context{
					Event.Address: httpRequest.RemoteAddr,
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
				"rejecting websocket upgrade request because server is stopping",
				server.GetServerContext().Merge(Event.Context{
					Event.Kind: Event.StopChannel,
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
				"sending new websocketConnection to channel",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:    Event.WebsocketConnection,
					Event.Address: websocketConnection.RemoteAddr().String(),
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
				"sent upgraded websocketConnection to channel",
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:    Event.WebsocketConnection,
					Event.Address: websocketConnection.RemoteAddr().String(),
				}),
			))
		}
	}
}
