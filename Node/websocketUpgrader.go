package Node

import (
	"Systemge/Error"
	"Systemge/Http"
	"net/http"
)

func (node *Node) WebsocketUpgrade() Http.RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if !node.isStarted {
			node.GetWarningLogger().Log(Error.New("websocket connection not accepted", nil).Error())
			return
		}
		if err := node.validateWebsocketAddress(httpRequest.RemoteAddr); err != nil {
			node.GetWarningLogger().Log(Error.New("websocket connection address not accepted", err).Error())
			return
		}
		websocketConn, err := node.GetWebsocketComponent().GetWebsocketComponentConfig().Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			node.GetWarningLogger().Log(Error.New("failed upgrading connection to websocket", err).Error())
			return
		}
		node.websocketConnChannel <- websocketConn
	}
}
