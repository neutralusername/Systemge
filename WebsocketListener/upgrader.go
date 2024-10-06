package WebsocketListener

import (
	"net/http"
)

func (server *WebsocketListener) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		select {
		case <-server.stopChannel:
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.ClientsRejected.Add(1)
			return
		case acceptRequest := <-server.acceptChannel:
			websocketConn, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
			if err != nil {
				server.ClientsFailed.Add(1)
			}
			acceptRequest.upgraderResponseChannel <- &upgraderResponse{
				err:           err,
				websocketConn: websocketConn,
			}
		}
	}
}
