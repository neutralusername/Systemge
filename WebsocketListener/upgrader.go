package WebsocketListener

import (
	"net/http"
)

func (server *WebsocketListener) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		for {
			select {
			case <-server.stopChannel:
				http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
				server.ClientsRejected.Add(1)
				return
			case acceptRequest := <-server.acceptChannel:
				if acceptRequest.timedOut {
					continue
				}

				websocketConn, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)

				acceptRequest.upgraderResponseChannel <- &upgraderResponse{
					err:           err,
					websocketConn: websocketConn,
				}
				acceptRequest.triggered.Wait()
				if err != nil {
					server.ClientsFailed.Add(1)
				} else {
					if acceptRequest.timedOut {
						websocketConn.Close()
						server.ClientsFailed.Add(1)
					}
				}
			}
		}
	}
}
