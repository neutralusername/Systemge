package WebsocketServer

import (
	"net"
	"net/http"

	"github.com/neutralusername/Systemge/Error"
)

func (server *WebsocketServer) getHTTPWebsocketUpgradeHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if server.httpServer == nil {
			if server.warningLogger != nil {
				server.warningLogger.Log(Error.New("websocket component not started", nil).Error())
			}
			return
		}
		ip, _, err := net.SplitHostPort(httpRequest.RemoteAddr)
		if err != nil {
			if server.warningLogger != nil {
				server.warningLogger.Log(Error.New("failed to split IP and port", err).Error())
			}
			http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
			return
		}
		if server.ipRateLimiter != nil && !server.ipRateLimiter.RegisterConnectionAttempt(ip) {
			if server.warningLogger != nil {
				server.warningLogger.Log("IP rate limit exceeded for " + httpRequest.RemoteAddr)
			}
			http.Error(responseWriter, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		websocketConnection, err := server.config.Upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			if server.warningLogger != nil {
				server.warningLogger.Log(Error.New("failed upgrading connection to websocket", err).Error())
			}
			return
		}
		server.connectionChannel <- websocketConnection
	}
}
