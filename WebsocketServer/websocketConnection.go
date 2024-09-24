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

	server *WebsocketServer

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
func (websocketConnection *WebsocketConnection) GetIp() string {
	return websocketConnection.websocketConn.RemoteAddr().String()
}

// Returns the id of the websocketConnection.
func (websocketConnection *WebsocketConnection) GetId() string {
	return websocketConnection.id
}

func (server *WebsocketServer) Send(websocketConnection *WebsocketConnection, messageBytes []byte) error {
	return server.send(websocketConnection, messageBytes, Event.SendRuntime)
}
func (server *WebsocketServer) send(websocketConnection *WebsocketConnection, messageBytes []byte, circumstance string) error {
	websocketConnection.sendMutex.Lock()
	defer websocketConnection.sendMutex.Unlock()

	if event := server.onEvent(Event.NewInfo(
		Event.SendingClientMessage,
		"sending websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:  circumstance,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetIp(),
			Event.Bytes:         string(messageBytes),
		}),
	)); !event.IsInfo() {
		server.websocketConnectionMessagesFailed.Add(1)
		return event.GetError()
	}

	err := websocketConnection.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		server.websocketConnectionMessagesFailed.Add(1)
		server.onEvent(Event.NewWarningNoOption(
			Event.SendingClientMessageFailed,
			err.Error(),
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:  circumstance,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetIp(),
				Event.Bytes:         string(messageBytes),
			}),
		))
		return err
	}
	server.websocketConnectionMessagesSent.Add(1)
	server.websocketConnectionMessagesBytesSent.Add(uint64(len(messageBytes)))

	server.onEvent(Event.NewInfoNoOption(
		Event.SentClientMessage,
		"sent websocketConnection message",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:  circumstance,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetIp(),
			Event.Bytes:         string(messageBytes),
		}),
	))
	return nil
}

func (server *WebsocketServer) Receive(websocketConnection *WebsocketConnection) ([]byte, error) {
	if websocketConnection.isAccepted {
		server.onEvent(Event.NewWarningNoOption(
			Event.ClientAlreadyAccepted,
			"websocketConnection is already accepted",
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:  Event.ReceiveRuntime,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetIp(),
			}),
		))
		return nil, errors.New("websocketConnection is already accepted")
	}
	return server.receive(websocketConnection, Event.ReceiveRuntime)
}
func (server *WebsocketServer) receive(websocketConnection *WebsocketConnection, circumstance string) ([]byte, error) {
	websocketConnection.receiveMutex.Lock()
	defer websocketConnection.receiveMutex.Unlock()

	if event := server.onEvent(Event.NewInfo(
		Event.ReceivingClientMessage,
		"receiving websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:  circumstance,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetIp(),
		}),
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	websocketConnection.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(server.config.ServerReadDeadlineMs) * time.Millisecond))
	_, messageBytes, err := websocketConnection.websocketConn.ReadMessage()
	if err != nil {
		websocketConnection.Close()
		server.onEvent(Event.NewWarningNoOption(
			Event.ReceivingClientMessageFailed,
			err.Error(),
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:  circumstance,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetIp(),
			}),
		))
		return nil, err
	}

	server.onEvent(Event.NewInfoNoOption(
		Event.ReceivedClientMessage,
		"received websocketConnection message",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:  circumstance,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetIp(),
			Event.Bytes:         string(messageBytes),
		}),
	))
	return messageBytes, nil
}
