package DashboardServer

import (
	"github.com/neutralusername/systemge/DashboardHelpers"
	"github.com/neutralusername/systemge/Message"
	"github.com/neutralusername/systemge/tools"
)

func (server *Server) handleWebsocketResponseMessage(responseMessage, page string) error {
	server.mutex.Lock()
	responseId := tools.GenerateRandomString(16, tools.ALPHA_NUMERIC)
	for server.responseMessageCache[responseId] != nil {
		responseId = tools.GenerateRandomString(16, tools.ALPHA_NUMERIC)
	}
	responseMessageStruct := DashboardHelpers.NewResponseMessage(responseId, page, responseMessage)
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
