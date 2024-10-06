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
				server.onEvent(Event.New(
					Event.WebsocketUpgradeFailed,
					Event.Context{
						Event.Address: httpRequest.RemoteAddr,
						Event.Error:   err.Error(),
					},
					Event.Continue,
				))
			}
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.clientsFailed.Add(1)
			return
		}

		if server.eventHandler != nil {
			if event := server.onEvent(Event.New(
				Event.SendingToChannel,
				Event.Context{
					Event.Address: websocketConn.RemoteAddr().String(),
				},
				Event.Continue,
				Event.Cancel,
			)); event.GetAction() == Event.Cancel {
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError) // idk if this will work after upgrade
				websocketConn.Close()
				server.clientsRejected.Add(1)
			}
		}

		server.connectionChannel <- websocketConn

		if server.eventHandler != nil {
			server.onEvent(Event.New(
				Event.SentToChannel,
				Event.Context{
					Event.ChannelType: Event.WebsocketClient,
					Event.Address:     websocketConn.RemoteAddr().String(),
				},
				Event.Continue,
			))
		}
	}
}
