package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (server *WebsocketServer[O]) receptionRoutine(session *Tools.Session, websocketClient *WebsocketClient.WebsocketClient) {
	defer func() {
		if server.eventHandler != nil {
			server.eventHandler.Handle(Event.New(
				Event.ReceptionRoutineEnds,
				Event.Context{},
				Event.Continue,
				Event.Cancel,
			))
		}
		server.waitGroup.Done()
	}()

	if server.eventHandler != nil {
		event := server.eventHandler.Handle(Event.New(
			Event.ReceptionRoutineBegins,
			Event.Context{},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			return
		}
	}

	websocketServerReceptionHandlerCaller := &websocketServerReceptionHandlerCaller{
		Client:    websocketClient,
		SessionId: session.GetId(),
		Identity:  session.GetIdentity(),
	}
	receptionHandler := server.receptionHandlerFactory()
	err := receptionHandler.Start(websocketServerReceptionHandlerCaller)
	if err != nil {
		// event
		// close connection?
		return
	}
	handleReceptionWrapper := func(session *Tools.Session, websocketClient *WebsocketClient.WebsocketClient, messageBytes []byte) {
		// event
		if err := receptionHandler.HandleReception(messageBytes, websocketServerReceptionHandlerCaller); err != nil {
			server.FailedReceptions.Add(1)
			if server.eventHandler != nil {
				event := server.eventHandler.Handle(Event.New(
					Event.ReceptionHandlerFailed,
					Event.Context{
						Event.SessionId: session.GetId(),
						Event.Identity:  session.GetIdentity(),
						Event.Address:   websocketClient.GetAddress(),
						Event.Error:     err.Error(),
					},
					Event.Skip,
					Event.Cancel,
				))
				if event.GetAction() == Event.Cancel {
					websocketClient.Close()
				}
			}
		} else {
			server.SuccessfulReceptions.Add(1)
		}
		// event
	}

	for {
		messageBytes, err := websocketClient.Read(server.config.ReadTimeoutMs)
		if err != nil {
			if server.eventHandler != nil {
				event := server.eventHandler.Handle(Event.New(
					Event.ReadMessageFailed,
					Event.Context{
						Event.SessionId: session.GetId(),
						Event.Identity:  session.GetIdentity(),
						Event.Address:   websocketClient.GetAddress(),
						Event.Error:     err.Error(),
					},
					Event.Cancel,
					Event.Skip,
				))
				if event.GetAction() == Event.Skip {
					continue
				}
			}
			websocketClient.Close()
			break
		}

		if server.config.HandleMessagesSequentially {
			handleReceptionWrapper(session, websocketClient, messageBytes)
		} else {
			server.waitGroup.Add(1)
			go func(websocketClient *WebsocketClient.WebsocketClient) {
				handleReceptionWrapper(session, websocketClient, messageBytes)
				server.waitGroup.Done()
			}(websocketClient)
		}
	}
}
