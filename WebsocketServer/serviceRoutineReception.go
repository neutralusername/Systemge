package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (server *WebsocketServer) receptionRoutine(session *Tools.Session, websocketClient *WebsocketClient.WebsocketClient) {
	defer func() {
		if server.eventHandler != nil {
			server.onEvent(Event.New(
				Event.ReceptionRoutineEnds,
				Event.Context{},
				Event.Continue,
				Event.Cancel,
			))
		}
		server.waitGroup.Done()
	}()

	if server.eventHandler != nil {
		event := server.onEvent(Event.New(
			Event.ReceptionRoutineBegins,
			Event.Context{},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			return
		}
	}

	handleReceptionWrapper := func(session *Tools.Session, websocketClient *WebsocketClient.WebsocketClient, messageBytes []byte) {
		if err := server.receptionHandler(websocketClient, messageBytes); err != nil {
			websocketClient.Close()
			server.RejectedMessages.Add(1)
		} else {
			server.AcceptedMessages.Add(1)
		}
	}

	for {
		messageBytes, err := websocketClient.Read(server.config.ReadTimeoutMs)
		if err != nil {
			if server.eventHandler != nil {
				event := server.onEvent(Event.New(
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
