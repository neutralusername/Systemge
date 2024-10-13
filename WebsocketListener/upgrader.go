package WebsocketListener

import (
	"net/http"
	"time"
)

func (listener *WebsocketListener) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		upgradeResponseChannel := make(chan *upgraderResponse)

		var requestDeadline <-chan time.Time
		if listener.config.UpgradeRequestTimeoutMs > 0 {
			requestDeadline = time.After(time.Duration(listener.config.UpgradeRequestTimeoutMs) * time.Millisecond)
		}

		select {
		case <-listener.stopChannel:
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			listener.ClientsRejected.Add(1)
			return

		case <-requestDeadline:
			http.Error(responseWriter, "Request timeout", http.StatusRequestTimeout)
			listener.ClientsRejected.Add(1)
			return

		case listener.upgadeRequests <- upgradeResponseChannel:
			websocketConn, err := listener.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
			select {
			case upgradeResponseChannel <- &upgraderResponse{
				err:           err,
				websocketConn: websocketConn,
			}:

			default:
				if websocketConn != nil {
					websocketConn.Close()
				}
			}
		}
	}
}
