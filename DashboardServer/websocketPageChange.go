package DashboardServer

import (
	"github.com/neutralusername/systemge/DashboardHelpers"
	"github.com/neutralusername/systemge/Event"
	"github.com/neutralusername/systemge/Message"
	"github.com/neutralusername/systemge/WebsocketServer"
)

func (server *Server) changePageHandler(websocketClient *WebsocketServer.WebsocketConnection, message *Message.Message) error {
	locationAfterChange := message.GetPayload()
	return server.changePage(websocketClient, locationAfterChange, true)
}
func (server *Server) changePage(websocketClient *WebsocketServer.WebsocketConnection, locationAfterChange string, lock bool) error {
	if lock {
		server.mutex.Lock()
		defer server.mutex.Unlock()
	}
	locationBeforeChange := server.websocketClientLocations[websocketClient]

	if locationBeforeChange == locationAfterChange {
		return Event.New("Location is already "+locationAfterChange, nil)
	}

	var pageJson string
	var connectedClient *connectedClient
	switch locationAfterChange {
	case "":
		page, err := DashboardHelpers.GetNullPage().Marshal()
		if err != nil {
			return Event.New("Failed to marshal null page", err)
		}
		pageJson = string(page)
	case DashboardHelpers.DASHBOARD_CLIENT_NAME:
		pageMarshalled, err := DashboardHelpers.NewPage(
			server.dashboardClient,
			DashboardHelpers.CLIENT_TYPE_DASHBOARD,
		).Marshal()
		if err != nil {
			return Event.New("Failed to marshal dashboard page", err)
		}
		pageJson = string(pageMarshalled)
		server.dashboardWebsocketClients[websocketClient] = true
	default:
		connectedClient = server.connectedClients[locationAfterChange]
		if connectedClient == nil {
			return Event.New("Client not found", nil)
		}
		pageMarshalled, err := connectedClient.page.Marshal()
		if err != nil {
			return Event.New("Failed to marshal client page", err)
		}
		pageJson = string(pageMarshalled)
		connectedClient.websocketClients[websocketClient] = true
	}
	server.websocketClientLocations[websocketClient] = locationAfterChange
	switch locationBeforeChange {
	case "":
	case DashboardHelpers.DASHBOARD_CLIENT_NAME:
		delete(server.dashboardWebsocketClients, websocketClient)
	default:
		if connectedClient != nil {
			delete(connectedClient.websocketClients, websocketClient)
		}
	}
	go websocketClient.Send(Message.NewAsync(
		DashboardHelpers.TOPIC_CHANGE_PAGE,
		pageJson,
	).Serialize())
	return nil
}
