package WebsocketServer

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

type WebsocketConnection struct {
	id            string
	websocketConn *websocket.Conn

	isAccepted bool

	receiveMutex sync.Mutex
	sendMutex    sync.Mutex
	stopChannel  chan bool
	waitGroup    sync.WaitGroup

	closeMutex sync.Mutex
	isClosed   bool

	rateLimiterBytes *Tools.TokenBucketRateLimiter
	rateLimiterMsgs  *Tools.TokenBucketRateLimiter
}

// Disconnects the websocketConnection and blocks until the websocketConnection onDisconnectHandler has finished.
func (websocketConnection *WebsocketConnection) Close() error {
	websocketConnection.closeMutex.Lock()
	defer websocketConnection.closeMutex.Unlock()
	if websocketConnection.isClosed {
		return errors.New("websocketConnection is already closed")
	}
	websocketConnection.isClosed = true
	websocketConnection.websocketConn.Close()
	if websocketConnection.rateLimiterBytes != nil {
		websocketConnection.rateLimiterBytes.Close()
	}
	if websocketConnection.rateLimiterMsgs != nil {
		websocketConnection.rateLimiterMsgs.Close()
	}
	close(websocketConnection.stopChannel)
	return nil
}

// Returns the ip of the websocketConnection.
func (websocketConnection *WebsocketConnection) GetAddress() string {
	return websocketConnection.websocketConn.RemoteAddr().String()
}

// Returns the id of the websocketConnection.
func (websocketConnection *WebsocketConnection) GetId() string {
	return websocketConnection.id
}

func (server *WebsocketServer) Send(websocketConnection *WebsocketConnection, messageBytes []byte) error {
	return server.write(websocketConnection, messageBytes, Event.SendRuntime)
}
func (server *WebsocketServer) write(websocketConnection *WebsocketConnection, messageBytes []byte, circumstance string) error {
	websocketConnection.sendMutex.Lock()
	defer websocketConnection.sendMutex.Unlock()

	if event := server.onEvent(Event.NewInfo(
		Event.WritingMessage,
		"sending websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: circumstance,
			Event.IdentityType: Event.WebsocketConnection,
			Event.Identity:     websocketConnection.GetId(),
			Event.Address:      websocketConnection.GetAddress(),
			Event.Bytes:        string(messageBytes),
		},
	)); !event.IsInfo() {
		server.websocketConnectionMessagesFailed.Add(1)
		return event.GetError()
	}

	err := websocketConnection.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		server.websocketConnectionMessagesFailed.Add(1)
		server.onEvent(Event.NewWarningNoOption(
			Event.WriteMessageFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: circumstance,
				Event.IdentityType: Event.WebsocketConnection,
				Event.Identity:     websocketConnection.GetId(),
				Event.Address:      websocketConnection.GetAddress(),
				Event.Bytes:        string(messageBytes),
			}),
		)
		return err
	}
	server.websocketConnectionMessagesSent.Add(1)
	server.websocketConnectionMessagesBytesSent.Add(uint64(len(messageBytes)))

	server.onEvent(Event.NewInfoNoOption(
		Event.WroteMessage,
		"sent websocketConnection message",
		Event.Context{
			Event.Circumstance: circumstance,
			Event.IdentityType: Event.WebsocketConnection,
			Event.Identity:     websocketConnection.GetId(),
			Event.Address:      websocketConnection.GetAddress(),
			Event.Bytes:        string(messageBytes),
		}),
	)
	return nil
}

func (server *WebsocketServer) Receive(websocketConnection *WebsocketConnection) ([]byte, error) {
	if websocketConnection.isAccepted {
		server.onEvent(Event.NewWarningNoOption(
			Event.ClientAlreadyAccepted,
			"websocketConnection is already accepted",
			Event.Context{
				Event.Circumstance: Event.ReceiveRuntime,
				Event.IdentityType: Event.WebsocketConnection,
				Event.Identity:     websocketConnection.GetId(),
				Event.Address:      websocketConnection.GetAddress(),
			}),
		)
		return nil, errors.New("websocketConnection is already accepted")
	}
	return server.receive(websocketConnection, Event.ReceiveRuntime)
}
func (server *WebsocketServer) receive(websocketConnection *WebsocketConnection, circumstance string) ([]byte, error) {
	websocketConnection.receiveMutex.Lock()
	defer websocketConnection.receiveMutex.Unlock()

	if event := server.onEvent(Event.NewInfo(
		Event.ReadingMessage,
		"receiving websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: circumstance,
			Event.IdentityType: Event.WebsocketConnection,
			Event.Identity:     websocketConnection.GetId(),
			Event.Address:      websocketConnection.GetAddress(),
		}),
	); !event.IsInfo() {
		return nil, event.GetError()
	}

	websocketConnection.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(server.config.ServerReadDeadlineMs) * time.Millisecond))
	_, messageBytes, err := websocketConnection.websocketConn.ReadMessage()
	if err != nil {
		websocketConnection.Close()
		return nil, err
	}

	if event := server.onEvent(Event.NewInfo(
		Event.ReadMessage,
		"received websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: circumstance,
			Event.IdentityType: Event.WebsocketConnection,
			Event.Identity:     websocketConnection.GetId(),
			Event.Address:      websocketConnection.GetAddress(),
			Event.Bytes:        string(messageBytes),
		}),
	); !event.IsInfo() {
		return nil, event.GetError()
	}
	return messageBytes, nil
}
