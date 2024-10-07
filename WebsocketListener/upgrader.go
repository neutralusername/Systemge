package WebsocketListener

import (
	"net/http"
)

func (listener *WebsocketListener) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		select {
		case <-listener.stopChannel:
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			listener.ClientsRejected.Add(1)
			return

		case acceptRequest, ok := <-listener.pool.AcquireItemChannel(listener.config.WebsocketRequestTimeoutMs):
			if !ok {
				http.Error(responseWriter, "Request timeout", http.StatusRequestTimeout)
				listener.ClientsRejected.Add(1)
				return
			}

			websocketConn, err := listener.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
			acceptRequest.upgraderResponseChannel <- &upgraderResponse{
				err:           err,
				websocketConn: websocketConn,
			}
			if err != nil {
				listener.ClientsFailed.Add(1)
			} else {
				acceptRequest.triggered.Wait()
				select {
				case <-acceptRequest.upgraderResponseChannel:
					websocketConn.Close()
					listener.ClientsFailed.Add(1)
				default:
				}
			}
		}
	}
}
