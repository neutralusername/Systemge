package WebsocketServer

import (
	"errors"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *WebsocketServer) acceptRoutine() {
	defer func() {
		server.onEvent(Event.NewInfoNoOption(
			Event.AcceptionRoutineFinished,
			"stopped websocketConnection acception routine",
			Event.Context{
				Event.Circumstance: Event.AcceptionRoutine,
				Event.ClientType:   Event.WebsocketConnection,
			},
		))
		server.waitGroup.Done()
	}()

	if event := server.onEvent(Event.NewInfo(
		Event.AcceptionRoutineStarted,
		"started websocketConnection acception routine",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AcceptionRoutine,
			Event.ClientType:   Event.WebsocketConnection,
		},
	)); !event.IsInfo() {
		return
	}

	for err := server.receiveWebsocketConnection(); err == nil; {
	}

}

func (server *WebsocketServer) receiveWebsocketConnection() error {
	if event := server.onEvent(Event.NewInfo(
		Event.ReceivingFromChannel,
		"receiving websocketConnection from channel",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AcceptionRoutine,
			Event.ChannelType:  Event.WebsocketConnection,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	websocketConnection := <-server.connectionChannel
	if websocketConnection == nil {
		server.onEvent(Event.NewInfoNoOption(
			Event.ReceivedNilValueFromChannel,
			"received nil from websocketConnection channel",
			Event.Context{
				Event.Circumstance: Event.AcceptionRoutine,
				Event.ChannelType:  Event.WebsocketConnection,
			},
		))
		return errors.New("received nil from websocketConnection channel")
	}

	event := server.onEvent(Event.NewInfo(
		Event.ReceivedFromChannel,
		"received websocketConnection from channel",
		Event.Cancel,
		Event.Skip,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.AcceptionRoutine,
			Event.ChannelType:   Event.WebsocketConnection,
			Event.ClientAddress: websocketConnection.RemoteAddr().String(),
		},
	))
	if event.IsError() {
		websocketConnection.Close()
		server.websocketConnectionsRejected.Add(1)
		return event.GetError()
	}
	if event.IsWarning() {
		websocketConnection.Close()
		server.websocketConnectionsRejected.Add(1)
		return nil
	}

	server.acceptWebsocketConnection(websocketConnection)
	return nil
}

func (server *WebsocketServer) acceptWebsocketConnection(websocketConn *websocket.Conn) {
	server.websocketConnectionMutex.Lock()

	websocketId := server.randomizer.GenerateRandomString(Constants.ClientIdLength, Tools.ALPHA_NUMERIC)
	for _, exists := server.websocketConnections[websocketId]; exists; {
		websocketId = server.randomizer.GenerateRandomString(Constants.ClientIdLength, Tools.ALPHA_NUMERIC)
	}
	if event := server.onEvent(Event.NewInfo(
		Event.HandlingAcception,
		"accepting websocketConnection",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.HandleAcception,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketId,
			Event.ClientAddress: websocketConn.RemoteAddr().String(),
		},
	)); !event.IsInfo() {
		websocketConn.Close()
		server.websocketConnectionsRejected.Add(1)
		return
	}

	websocketConnection := &WebsocketConnection{
		id:            websocketId,
		websocketConn: websocketConn,
		stopChannel:   make(chan bool),
	}
	if server.config.WebsocketConnectionRateLimiterBytes != nil {
		websocketConnection.rateLimiterBytes = Tools.NewTokenBucketRateLimiter(server.config.WebsocketConnectionRateLimiterBytes)
	}
	if server.config.WebsocketConnectionRateLimiterMessages != nil {
		websocketConnection.rateLimiterMsgs = Tools.NewTokenBucketRateLimiter(server.config.WebsocketConnectionRateLimiterMessages)
	}
	websocketConnection.websocketConn.SetReadLimit(int64(server.config.IncomingMessageByteLimit))
	server.websocketConnections[websocketId] = websocketConnection
	server.websocketConnectionGroups[websocketId] = make(map[string]bool)
	server.websocketConnectionMutex.Unlock()

	if event := server.onEvent(Event.NewInfo(
		Event.HandledAcception,
		"websocketConnection accepted",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.HandleAcception,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketId,
			Event.ClientAddress: websocketConn.RemoteAddr().String(),
		},
	)); !event.IsInfo() {
		websocketConnection.Close()
		server.removeWebsocketConnection(websocketConnection)
		server.websocketConnectionsRejected.Add(1)
		return
	}

	server.websocketConnectionsAccepted.Add(1)
	websocketConnection.isAccepted = true

	websocketConnection.waitGroup.Add(1)
	server.waitGroup.Add(1)
	go server.websocketConnectionDisconnect(websocketConnection)
	go server.receptionRoutine(websocketConnection)
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
		Event.HandlingDisconnection,
		"disconnecting websocketConnection",
		Event.Context{
			Event.Circumstance:  Event.HandleDisconnection,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetAddress(),
		},
	))
	server.removeWebsocketConnection(websocketConnection)

	server.onEvent(Event.NewInfoNoOption(
		Event.HandledDisconnection,
		"websocketConnection disconnected",
		Event.Context{
			Event.Circumstance:  Event.HandleDisconnection,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetAddress(),
		},
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
