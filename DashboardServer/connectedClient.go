package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

type connectedClient struct {
	connection       SystemgeConnection.SystemgeConnection
	websocketClients map[*WebsocketServer.WebsocketClient]bool // websocketClient -> true (websocketClients that are currently on this client's page)
	client           interface{}
}

func newConnectedClient(connection SystemgeConnection.SystemgeConnection, client interface{}) *connectedClient {
	return &connectedClient{
		connection:       connection,
		websocketClients: make(map[*WebsocketServer.WebsocketClient]bool),
		client:           client,
	}
}

func (connectedClient *connectedClient) executeCommand(command string, args []string) (string, error) {
	response, err := connectedClient.connection.SyncRequestBlocking(
		Message.TOPIC_EXECUTE_COMMAND,
		DashboardHelpers.NewCommand(
			command,
			args,
		).Marshal(),
	)
	if err != nil {
		return "", Error.New("Failed to send command \""+command+"\" to client \""+connectedClient.connection.GetName()+"\"", err)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return "", Error.New(response.GetPayload(), nil)
	}
	return response.GetPayload(), nil
}

func (connectedClient *connectedClient) executeStart() (int, error) {
	response, err := connectedClient.connection.SyncRequestBlocking(
		Message.TOPIC_START,
		"",
	)
	if err != nil {
		return Status.NON_EXISTENT, Error.New("Failed to send start request to client \""+connectedClient.connection.GetName()+"\": "+err.Error(), nil)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return Status.NON_EXISTENT, Error.New(response.GetPayload(), nil)
	}
	return Helpers.StringToInt(response.GetPayload()), nil
}

func (connectedClient *connectedClient) executeStop() (int, error) {
	response, err := connectedClient.connection.SyncRequestBlocking(
		Message.TOPIC_STOP,
		"",
	)
	if err != nil {
		return Status.NON_EXISTENT, Error.New("Failed to send stop request to client \""+connectedClient.connection.GetName()+"\": "+err.Error(), nil)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return Status.NON_EXISTENT, Error.New(response.GetPayload(), nil)
	}
	return Helpers.StringToInt(response.GetPayload()), nil
}
