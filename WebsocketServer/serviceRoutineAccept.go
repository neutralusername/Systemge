package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/ReceptionHandler"
	"github.com/neutralusername/Systemge/SessionManager"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *WebsocketServer) sessionRoutine() {
	defer func() {
		server.onEvent___(Event.NewInfoNoOption(
			Event.SessionRoutineEnds,
			"stopped websocketConnection session routine",
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
			},
		))
		server.waitGroup.Done()
	}()

	if event := server.onEvent___(Event.NewInfo(
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

	for err := server.handleNewSession(); err == nil; {
	}

}

func (server *WebsocketServer) handleNewSession() error {
	websocketConn := <-server.connectionChannel
	if websocketConn == nil {
		server.onEvent___(Event.NewInfoNoOption(
			Event.ReceivedNilValueFromChannel,
			"received nil from websocketConnection channel",
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
				Event.ChannelType:  Event.WebsocketConnection,
			},
		))
		return errors.New("received nil from websocketConnection channel")
	}

	websocketConnection := server.NewWebsocketConnection(websocketConn)
	session, err := server.sessionManager.CreateSession("", map[string]any{
		"connection": websocketConnection,
	})
	if err != nil {
		server.onEvent___(Event.NewWarningNoOption(
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
		return nil
	}
	if server.config.RateLimiterBytes != nil {
		websocketConnection.byteRateLimiter = Tools.NewTokenBucketRateLimiter(server.config.RateLimiterBytes)
	}
	if server.config.RateLimiterMessages != nil {
		websocketConnection.messageRateLimiter = Tools.NewTokenBucketRateLimiter(server.config.RateLimiterMessages)
	}
	pipeline, _ := ReceptionHandler.NewReceptionHandler(session.GetId()+"_pipeline", websocketConnection.byteRateLimiter, websocketConnection.messageRateLimiter, server.deserializer, server.validator, server.eventHandler)
	websocketConnection.pipeline = pipeline
	websocketConnection.id = session.GetId()

	server.websocketConnectionsAccepted.Add(1)
	websocketConnection.waitGroup.Add(1)
	server.waitGroup.Add(1)
	go server.websocketConnectionDisconnect(session, websocketConnection)
	go server.receptionRoutine(websocketConnection)

	return nil
}

func (server *WebsocketServer) websocketConnectionDisconnect(session *SessionManager.Session, websocketConnection *WebsocketConnection) {
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
