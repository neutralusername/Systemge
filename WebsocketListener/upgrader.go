package WebsocketListener

import (
	"net/http"

	"github.com/neutralusername/Systemge/Event"
)

func (server *WebsocketListener) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		websocketConn, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			if server.eventHandler != nil {
				server.onEvent(Event.NewWarningNoOption(
					Event.WebsocketUpgradeFailed,
					err.Error(),
					Event.Context{
						Event.Circumstance: Event.WebsocketUpgrade,
						Event.Address:      httpRequest.RemoteAddr,
					}),
				)
			}
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.failed.Add(1)
			return
		}
		if server.eventHandler != nil {
			if event := server.onEvent(Event.NewInfo(
				Event.SendingToChannel,
				"sending websocketConn to channel",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance: Event.WebsocketUpgrade,
					Event.Address:      websocketConn.RemoteAddr().String(),
				}),
			); !event.IsInfo() {
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError) // idk if this will work after upgrade
				websocketConn.Close()
				server.rejected.Add(1)
			}
		}
		server.connectionChannel <- websocketConn
		if server.eventHandler != nil {
			server.onEvent(Event.NewInfoNoOption(
				Event.SentToChannel,
				"sent websocketConn to channel",
				Event.Context{
					Event.Circumstance: Event.WebsocketUpgrade,
					Event.ChannelType:  Event.WebsocketClient,
					Event.Address:      websocketConn.RemoteAddr().String(),
				}),
			)
		}
	}
}
