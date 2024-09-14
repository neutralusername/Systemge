package DashboardServer

import (
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
