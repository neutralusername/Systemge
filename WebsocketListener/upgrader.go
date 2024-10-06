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

				acceptRequest.mutex.Lock()
				if acceptRequest.timedOut {
					acceptRequest.mutex.Unlock()
					continue
				}

				websocketConn, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)

				// timeout branch could trigger here but for whatever (scheduler) reason, the goroutine does not progress until the mutex.Lock()
				acceptRequest.upgraderResponseChannel <- &upgraderResponse{
					err:           err,
					websocketConn: websocketConn,
				}
				if err != nil {
					server.ClientsFailed.Add(1)
				} else {
					// race conditiion (timeout branch could already be triggered here but did not set timeedOut to true yet ..which would cause the wsConn to stay open and unreachable)
					acceptRequest.mutex.Lock()
					timedOut := acceptRequest.timedOut
					acceptRequest.mutex.Unlock()
					if timedOut {
						websocketConn.Close()
						server.ClientsFailed.Add(1)
					}
				}
			}
		}
	}
}
