package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

func (server *Server) onWebsocketConnectHandler(websocketClient *WebsocketServer.WebsocketClient) error {
	if server.config.FrontendPassword != "" {
		err := websocketClient.Send(Message.NewAsync(
			DashboardHelpers.TOPIC_PASSWORD,
			"",
		).Serialize())
		if err != nil {
			return Error.New("Failed to send password request", err)
		}
		messageBytes, err := websocketClient.Receive()
		if err != nil {
			return Error.New("Failed to receive message", err)
		}
		message, err := Message.Deserialize(messageBytes, websocketClient.GetId())
		if err != nil {
			return Error.New("Failed to deserialize message", err)
		}
		if message.GetTopic() != DashboardHelpers.TOPIC_PASSWORD {
			return Error.New("Expected password request", nil)
		}
		if message.GetPayload() != server.config.FrontendPassword {
			return Error.New("Invalid password", nil)
		}
	}

	err := websocketClient.Send(
		Message.NewAsync(
			DashboardHelpers.TOPIC_GET_RESPONSE_MESSAGE_CACHE,
			Helpers.JsonMarshal(server.responseMessageCacheOrder),
		).Serialize(),
	)
	if err != nil {
		return Error.New("Failed to send response message cache request", err)
	}

	err = websocketClient.Send(Message.NewAsync(
		DashboardHelpers.TOPIC_REQUEST_PAGE_CHANGE,
		"",
	).Serialize())
	if err != nil {
		return Error.New("Failed to send page change request", err)
	}

	server.mutex.Lock()
	defer server.mutex.Unlock()
	server.websocketClientLocations[websocketClient] = ""
	return nil
}
func (server *Server) onWebsocketDisconnectHandler(websocketClient *WebsocketServer.WebsocketClient) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	delete(server.websocketClientLocations, websocketClient)
}
