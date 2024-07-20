package Node

import (
	"Systemge/Error"
	"net/http"
)

func (node *Node) WebsocketUpgrade() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if !node.isStarted {
			node.GetWarningLogger().Log(Error.New("websocket connection not accepted", nil).Error())
			return
		}
		websocketConn, err := node.GetWebsocketComponent().GetWebsocketComponentConfig().Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("failed upgrading connection to websocket", err).Error())
			}
			return
		}
		node.websocketConnChannel <- websocketConn
	}
}
