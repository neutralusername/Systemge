package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (app *Server) onSystemgeConnectHandler(connection SystemgeConnection.SystemgeConnection) error {
	response, err := connection.SyncRequestBlocking(Message.TOPIC_INTRODUCTION, "")
	if err != nil {
		return err
	}
	client, err := DashboardHelpers.UnmarshalIntroduction([]byte(response.GetPayload()))
	if err != nil {
		return err
	}
	connectedClient := &connectedClient{
		connection: connection,
		client:     client,
	}

	app.mutex.Lock()
	app.registerModuleHttpHandlers(connectedClient)
	app.connectedClients[connection.GetName()] = connectedClient
	app.mutex.Unlock()

	app.websocketServer.Broadcast(Message.NewAsync("addModule", Helpers.JsonMarshal(client)))
	return nil
}

func (app *Server) onSystemgeDisconnectHandler(connection SystemgeConnection.SystemgeConnection) {
	app.mutex.Lock()
	if _, ok := app.connectedClients[connection.GetName()]; ok {
		delete(app.connectedClients, connection.GetName())
		app.unregisterModuleHttpHandlers(connection.GetName())
	}
	app.mutex.Unlock()
	app.websocketServer.Broadcast(Message.NewAsync("removeModule", connection.GetName()))
}
