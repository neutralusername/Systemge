package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

func (server *Server) handleWebsocketResponseMessage(websocketClient *WebsocketServer.WebsocketClient, responseMessage string) error {
	return websocketClient.Send(Message.NewAsync(
		DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
		responseMessage,
	).Serialize())
}
