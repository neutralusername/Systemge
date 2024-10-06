package WebsocketListener

import (
	"net/http"
	"time"
)

func (server *WebsocketListener) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		var deadline <-chan time.Time
		if server.config.WebsocketRequestTimeoutMs > 0 {
			deadline = time.After(time.Duration(server.config.WebsocketRequestTimeoutMs) * time.Millisecond)
		}
		select {
		case <-deadline:
			http.Error(responseWriter, "Request timeout", http.StatusRequestTimeout)
			server.ClientsRejected.Add(1)
			return

		case <-server.stopChannel:
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			server.ClientsRejected.Add(1)
			return

		case acceptRequest := <-server.pool.AcquireItemChannel():
			websocketConn, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)

			acceptRequest.upgraderResponseChannel <- &upgraderResponse{
				err:           err,
				websocketConn: websocketConn,
			}
			if err != nil {
				server.ClientsFailed.Add(1)
			} else {
				acceptRequest.triggered.Wait()
				select {
				case <-acceptRequest.upgraderResponseChannel:
					websocketConn.Close()
					server.ClientsFailed.Add(1)
				default:
				}
			}
		}
	}
}
