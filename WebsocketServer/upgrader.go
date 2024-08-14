package WebsocketServer

import (
	"net/http"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *WebsocketServer) getHTTPWebsocketUpgradeHandler(logger *Tools.Logger) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if server.httpServer == nil {
			if logger != nil {
				logger.Log(Error.New("websocket component not started", nil).Error())
			}
			return
		}
		websocketConnection, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			if logger != nil {
				logger.Log(Error.New("failed upgrading connection to websocket", err).Error())
			}
			return
		}
		server.connectionChannel <- websocketConnection
	}
}
