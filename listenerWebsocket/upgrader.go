package listenerWebsocket

import (
	"net/http"

	"github.com/neutralusername/systemge/tools"
)

func (listener *WebsocketListener) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		upgradeResponseChannel := make(chan *upgraderResponse)

		timeout := tools.NewTimeout(listener.config.UpgradeRequestTimeoutNs, nil, false)
		select {
		case <-listener.stopChannel:
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			listener.ClientsRejected.Add(1)
			return

		case <-timeout.GetIsExpiredChannel():
			http.Error(responseWriter, "Request timeout", http.StatusRequestTimeout)
			listener.ClientsRejected.Add(1)
			return

		case listener.upgradeRequests <- upgradeResponseChannel:

			websocketConn, err := listener.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)

			upgradeResponseChannel <- &upgraderResponse{
				err:           err,
				websocketConn: websocketConn,
			}
		}
	}
}
