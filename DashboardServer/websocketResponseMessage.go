package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *Server) handleWebsocketResponseMessage(responseMessage, page string) error {
	server.mutex.Lock()
	responseId := Tools.GenerateRandomString(16, Tools.ALPHA_NUMERIC)
	for server.responseMessageCache[responseId] != nil {
		responseId = Tools.GenerateRandomString(16, Tools.ALPHA_NUMERIC)
	}
	responseMessageStruct := DashboardHelpers.NewResponseMessage(responseId, responseMessage, page)
	server.responseMessageCache[responseId] = responseMessageStruct
	server.responseMessageCacheOrder = append(server.responseMessageCacheOrder, responseMessageStruct)
	if len(server.responseMessageCacheOrder) > server.config.ResponseMessageCacheSize {
		lastResponseMessage := server.responseMessageCacheOrder[0]
		delete(server.responseMessageCache, lastResponseMessage.Id)
		server.responseMessageCacheOrder = server.responseMessageCacheOrder[1:]
	}
	server.mutex.Unlock()

	server.websocketServer.Broadcast(
		Message.NewAsync(
			DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
			responseMessageStruct.Marshal(),
		),
	)
	return nil
}
