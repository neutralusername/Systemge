package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tools"
	"net/http"
)

func (websocket *websocketComponent) websocketUpgrade(logger *Tools.Logger) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if websocket.httpServer == nil {
			if logger != nil {
				logger.Log(Error.New("websocket component not started", nil).Error())
			}
			return
		}
		websocketConn, err := websocket.application.GetWebsocketComponentConfig().Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			if logger != nil {
				logger.Log(Error.New("failed upgrading connection to websocket", err).Error())
			}
			return
		}
		websocket.connChannel <- websocketConn
	}
}
