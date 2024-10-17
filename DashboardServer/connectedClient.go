package DashboardServer

import (
	"github.com/neutralusername/systemge/DashboardHelpers"
	"github.com/neutralusername/systemge/Event"
	"github.com/neutralusername/systemge/Message"
	"github.com/neutralusername/systemge/SystemgeConnection"
	"github.com/neutralusername/systemge/WebsocketServer"
)

type connectedClient struct {
	connection       SystemgeConnection.SystemgeConnection
	websocketClients map[*WebsocketServer.WebsocketConnection]bool // websocketClient -> true (websocketClients that are currently on this client's page)
	page             *DashboardHelpers.Page
}

func newConnectedClient(connection SystemgeConnection.SystemgeConnection, page *DashboardHelpers.Page) *connectedClient {
	return &connectedClient{
		connection:       connection,
		websocketClients: make(map[*WebsocketServer.WebsocketConnection]bool),
		page:             page,
	}
}

func (connectedClient *connectedClient) executeRequest(topic, payload string) (string, error) {
	response, err := connectedClient.connection.SyncRequestBlocking(
		topic,
		payload,
	)
	if err != nil {
		return "", Event.New("Failed to execute request", err)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return "", Event.New(response.GetPayload(), nil)
	}
	return response.GetPayload(), nil
}
