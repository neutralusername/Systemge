package WebsocketServer

import (
	"net/http"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tools"
)

func (websocketServer *WebsocketServer) websocketUpgrade(logger *Tools.Logger) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if websocketServer.httpServer == nil {
			if logger != nil {
				logger.Log(Error.New("websocket component not started", nil).Error())
			}
			return
		}
		websocketConn, err := websocketServer.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			if logger != nil {
				logger.Log(Error.New("failed upgrading connection to websocket", err).Error())
			}
			return
		}
		websocketServer.connChannel <- websocketConn
	}
}
