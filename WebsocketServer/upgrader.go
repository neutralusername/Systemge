package WebsocketServer

import (
	"net/http"

	"github.com/neutralusername/Systemge/Error"
)

func (server *WebsocketServer) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if server.httpServer == nil {
			if server.warningLogger != nil {
				server.warningLogger.Log(Error.New("websocket component not started", nil).Error())
			}
			return
		}
		websocketConnection, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			if server.warningLogger != nil {
				server.warningLogger.Log(Error.New("failed upgrading connection to websocket", err).Error())
			}
			return
		}
		server.connectionChannel <- websocketConnection
	}
}
