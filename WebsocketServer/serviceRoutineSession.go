package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *WebsocketServer) sessionRoutine() {
	defer func() {
		server.onEvent(Event.NewInfoNoOption(
			Event.SessionRoutineEnds,
			"stopped websocketConnection session routine",
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
			},
		))
		server.waitGroup.Done()
	}()

	if event := server.onEvent(Event.NewInfo(
		Event.SessionRoutineBegins,
		"started websocketConnection session routine",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.SessionRoutine,
		},
	)); !event.IsInfo() {
		return
	}

	for {
		event := server.onEvent(Event.NewInfo(
			Event.ReceivingFromChannel,
			"receiving websocketConnection from channel",
			Event.Cancel,
			Event.Skip,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
				Event.ChannelType:  Event.WebsocketConnection,
			},
		))
		if event.IsWarning() {
			continue
		}
		if event.IsError() {
			break
		}

		websocketConn := <-server.connectionChannel
		if websocketConn == nil {
			server.onEvent(Event.NewInfoNoOption(
				Event.ReceivedNilValueFromChannel,
				"received nil from websocketConnection channel",
				Event.Context{
					Event.Circumstance: Event.SessionRoutine,
					Event.ChannelType:  Event.WebsocketConnection,
				},
			))
			break
		}

		event = server.onEvent(Event.NewInfo(
			Event.ReceivedFromChannel,
			"received websocketConnection from channel",
			Event.Cancel,
			Event.Skip,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
				Event.ChannelType:  Event.WebsocketConnection,
				Event.Address:      websocketConn.RemoteAddr().String(),
			},
		))
		if event.IsWarning() {
			continue
		}
		if event.IsError() {
			break
		}

		websocketConnection := server.NewWebsocketConnection(websocketConn)
		session, err := server.sessionManager.CreateSession("", map[string]any{
			"connection": websocketConnection,
		})
		if err != nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.CreateSessionFailed,
				err.Error(),
				Event.Context{
					Event.Circumstance: Event.SessionRoutine,
					Event.Identity:     "",
					Event.Address:      websocketConnection.GetAddress(),
				},
			))
			server.websocketConnectionsRejected.Add(1)
			websocketConn.Close()
			continue
		}

		if server.config.RateLimiterBytes != nil {
			websocketConnection.byteRateLimiter = Tools.NewTokenBucketRateLimiter(server.config.RateLimiterBytes)
		}
		if server.config.RateLimiterMessages != nil {
			websocketConnection.messageRateLimiter = Tools.NewTokenBucketRateLimiter(server.config.RateLimiterMessages)
		}
		websocketConnection.id = session.GetId()

		server.websocketConnectionsAccepted.Add(1)
		websocketConnection.waitGroup.Add(1)
		server.waitGroup.Add(1)
		go server.websocketConnectionDisconnect(session, websocketConnection)
		go server.messageReceptionRoutine(websocketConnection)
	}
}

func (server *WebsocketServer) websocketConnectionDisconnect(session *Tools.Session, websocketConnection *WebsocketConnection) {
	select {
	case <-websocketConnection.stopChannel:
	case <-session.GetTimeout().GetTriggeredChannel():
	case <-server.stopChannel:
	}

	session.GetTimeout().Trigger()
	websocketConnection.Close()

	websocketConnection.waitGroup.Wait()
	server.waitGroup.Done()
}
