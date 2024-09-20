package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

func (server *Server) handleWebsocketResponseMessage(websocketClient *WebsocketServer.WebsocketClient, responseMessage string) error {
	server.mutex.Lock()
	server.responseMessageCache = append(server.responseMessageCache, DashboardHelpers.NewResponseMessage(responseMessage))
	if len(server.responseMessageCache) > server.config.ResponseMessageCacheSize {
		server.responseMessageCache = server.responseMessageCache[1:]
	}
	server.mutex.Unlock()

	return websocketClient.Send(Message.NewAsync(
		DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
		responseMessage,
	).Serialize())
}
