package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

type connectedClient struct {
	connection SystemgeConnection.SystemgeConnection
	client     interface{}
}

func (connectedClient *connectedClient) ExecuteCommand(command string, args []string) (string, error) {
	response, err := connectedClient.connection.SyncRequestBlocking(
		Message.TOPIC_EXECUTE_COMMAND,
		Helpers.JsonMarshal(
			&DashboardHelpers.Command{
				Command: command,
				Args:    args,
			},
		),
	)
	if err != nil {
		return "", Error.New("Failed to send command \""+command+"\" to client \""+connectedClient.connection.GetName()+"\"", err)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return "", Error.New(response.GetPayload(), nil)
	}
	return response.GetPayload(), nil
}
