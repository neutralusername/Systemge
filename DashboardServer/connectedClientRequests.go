package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

func (Server *Server) handleClientCommandRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message, connectedClient *connectedClient) error {
	command, err := DashboardHelpers.UnmarshalCommand(request.GetPayload())
	if err != nil {
		return Error.New("Failed to parse command", err)
	}
	if connectedClient == nil {
		return Error.New("Client not found", nil)
	}
	resultPayload, err := connectedClient.executeCommand(command.Command, command.Args)
	if err != nil {
		return Error.New("Failed to execute command", err)
	}
	websocketClient.Send(Message.NewAsync(
		DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
		resultPayload,
	).Serialize())
	return nil
}
