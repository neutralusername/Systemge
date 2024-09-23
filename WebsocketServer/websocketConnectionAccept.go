package WebsocketServer

import (
	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *WebsocketServer) acceptWebsocketConnection(websocketConn *websocket.Conn) {
	server.websocketConnectionMutex.Lock()

	websocketId := server.randomizer.GenerateRandomString(Constants.ClientIdLength, Tools.ALPHA_NUMERIC)
	for _, exists := server.websocketConnections[websocketId]; exists; {
		websocketId = server.randomizer.GenerateRandomString(Constants.ClientIdLength, Tools.ALPHA_NUMERIC)
	}
	if event := server.onEvent(Event.NewInfo(
		Event.AcceptingClient,
		"accepting websocketConnection",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:  Event.ClientAcceptionRoutine,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketId,
			Event.ClientAddress: websocketConn.RemoteAddr().String(),
		}),
	)); !event.IsInfo() {
		if websocketConn != nil {
			websocketConn.Close()
			server.websocketConnectionsRejected.Add(1)
		}
		return
	}

	websocketConnection := &WebsocketConnection{
		id:                  websocketId,
		websocketConnection: websocketConn,
		stopChannel:         make(chan bool),
	}
	if server.config.WebsocketConnectionRateLimiterBytes != nil {
		websocketConnection.rateLimiterBytes = Tools.NewTokenBucketRateLimiter(server.config.WebsocketConnectionRateLimiterBytes)
	}
	if server.config.WebsocketConnectionRateLimiterMessages != nil {
		websocketConnection.rateLimiterMsgs = Tools.NewTokenBucketRateLimiter(server.config.WebsocketConnectionRateLimiterMessages)
	}
	websocketConnection.websocketConnection.SetReadLimit(int64(server.config.IncomingMessageByteLimit))
	server.websocketConnections[websocketId] = websocketConnection
	server.websocketConnectionGroups[websocketId] = make(map[string]bool)
	server.websocketConnectionMutex.Unlock()

	websocketConnection.waitGroup.Add(1)
	server.waitGroup.Add(1)
	go server.websocketConnectionDisconnect(websocketConnection)

	if event := server.onEvent(Event.NewInfo(
		Event.AcceptedClient,
		"websocketConnection accepted",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:  Event.ClientAcceptionRoutine,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketId,
			Event.ClientAddress: websocketConn.RemoteAddr().String(),
		}),
	)); !event.IsInfo() {
		websocketConnection.Close()
		websocketConnection.waitGroup.Done()
		return
	}

	server.websocketConnectionsAccepted.Add(1)
	websocketConnection.isAccepted = true

	go server.receiveMessagesLoop(websocketConnection)
}

func (server *WebsocketServer) websocketConnectionDisconnect(websocketConnection *WebsocketConnection) {
	defer server.waitGroup.Done()

	select {
	case <-websocketConnection.stopChannel:
	case <-server.stopChannel:
		websocketConnection.Close()
	}

	websocketConnection.waitGroup.Wait()

	server.onEvent(Event.NewInfoNoOption(
		Event.DisconnectingClient,
		"disconnecting websocketConnection",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:  Event.ClientDisconnectionRoutine,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetIp(),
		}),
	))
	server.removeWebsocketConnection(websocketConnection)
	server.onEvent(Event.NewInfoNoOption(
		Event.DisconnectedClient,
		"websocketConnection disconnected",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:  Event.ClientDisconnectionRoutine,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetIp(),
		}),
	))
}
func (server *WebsocketServer) removeWebsocketConnection(websocketConnection *WebsocketConnection) {
	server.websocketConnectionMutex.Lock()
	defer server.websocketConnectionMutex.Unlock()
	delete(server.websocketConnections, websocketConnection.GetId())
	for groupId := range server.websocketConnectionGroups[websocketConnection.GetId()] {
		delete(server.websocketConnectionGroups[websocketConnection.GetId()], groupId)
		delete(server.groupsWebsocketConnections[groupId], websocketConnection.GetId())
		if len(server.groupsWebsocketConnections[groupId]) == 0 {
			delete(server.groupsWebsocketConnections, groupId)
		}
	}
}
