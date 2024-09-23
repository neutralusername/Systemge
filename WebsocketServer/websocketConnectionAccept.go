package WebsocketServer

import (
	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *WebsocketServer) acceptWebsocketConnection(websocketConn *websocket.Conn) {
	defer server.waitGroup.Done()

	if event := server.onInfo(Event.NewInfo(
		Event.AcceptingClient,
		"accepting websocketConnection",
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:    Event.WebsocketConnection,
			Event.Address: websocketConn.RemoteAddr().String(),
		}),
	)); event.IsError() {
		if websocketConn != nil {
			websocketConn.Close()
			server.websocketConnectionsRejected.Add(1)
		}
		return
	}

	server.websocketConnectionMutex.Lock()
	websocketId := server.randomizer.GenerateRandomString(16, Tools.ALPHA_NUMERIC)
	for _, exists := server.websocketConnections[websocketId]; exists; {
		websocketId = server.randomizer.GenerateRandomString(16, Tools.ALPHA_NUMERIC)
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
	go func() {
		defer server.waitGroup.Done()

		select {
		case <-websocketConnection.stopChannel:
		case <-server.stopChannel:
			websocketConnection.Close()
		}

		websocketConnection.waitGroup.Wait()

		server.onInfo(Event.NewInfoNoOption(
			Event.DisconnectingClient,
			"disconnecting websocketConnection",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:        Event.WebsocketConnection,
				Event.Address:     websocketConnection.GetIp(),
				Event.WebsocketId: websocketConnection.GetId(),
			}),
		))
		server.removeWebsocketConnection(websocketConnection)
		server.onInfo(Event.NewInfoNoOption(
			Event.DisconnectedClient,
			"websocketConnection disconnected",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:        Event.WebsocketConnection,
				Event.Address:     websocketConnection.GetIp(),
				Event.WebsocketId: websocketConnection.GetId(),
			}),
		))
	}()

	if event := server.onInfo(Event.NewInfo(
		Event.AcceptedClient,
		"websocketConnection accepted",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:        Event.WebsocketConnection,
			Event.Address:     websocketConn.RemoteAddr().String(),
			Event.WebsocketId: websocketId,
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
