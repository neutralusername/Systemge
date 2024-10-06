package WebsocketListener

import (
	"net/http"
	"time"
)

func (listener *WebsocketListener) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		var deadline <-chan time.Time
		if listener.config.WebsocketRequestTimeoutMs > 0 {
			deadline = time.After(time.Duration(listener.config.WebsocketRequestTimeoutMs) * time.Millisecond)
		}
		select {
		case <-deadline:
			http.Error(responseWriter, "Request timeout", http.StatusRequestTimeout)
			listener.ClientsRejected.Add(1)
			return

		case <-listener.stopChannel:
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			listener.ClientsRejected.Add(1)
			return

		case acceptRequest := <-listener.pool.AcquireItemChannel():
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
