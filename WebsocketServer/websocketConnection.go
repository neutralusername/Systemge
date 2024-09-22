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
	id                  string
	websocketConnection *websocket.Conn

	isAccepted bool

	receiveMutex sync.Mutex
	sendMutex    sync.Mutex
	stopChannel  chan bool

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
	websocketConnection.websocketConnection.Close()
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
	return websocketConnection.websocketConnection.RemoteAddr().String()
}

// Returns the id of the websocketConnection.
func (websocketConnection *WebsocketConnection) GetId() string {
	return websocketConnection.id
}

// Sends a message to the websocketConnection.
func (server *WebsocketServer) Send(websocketConnection *WebsocketConnection, messageBytes []byte) error {
	websocketConnection.sendMutex.Lock()
	defer websocketConnection.sendMutex.Unlock()

	if event := server.onInfo(Event.NewInfo(
		Event.SendingMessage,
		"sending message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:              Event.WebsocketConnection,
			Event.Address:           websocketConnection.GetIp(),
			Event.TargetWebsocketId: websocketConnection.GetId(),
			Event.Bytes:             string(messageBytes),
		}),
	)); !event.IsInfo() {
		server.failedSendCounter.Add(1)
		return event.GetError()
	}

	err := websocketConnection.websocketConnection.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		server.failedSendCounter.Add(1)
		server.onWarning(Event.NewWarningNoOption(
			Event.NetworkError,
			err.Error(),
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:              Event.WebsocketConnection,
				Event.Address:           websocketConnection.GetIp(),
				Event.TargetWebsocketId: websocketConnection.GetId(),
				Event.Bytes:             string(messageBytes),
			}),
		))
		return err
	}
	server.outgoigMessageCounter.Add(1)
	server.bytesSentCounter.Add(uint64(len(messageBytes)))

	server.onInfo(Event.NewInfoNoOption(
		Event.SentMessage,
		"sent message",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:              Event.WebsocketConnection,
			Event.Address:           websocketConnection.GetIp(),
			Event.TargetWebsocketId: websocketConnection.GetId(),
			Event.Bytes:             string(messageBytes),
		}),
	))
	return nil
}

func (server *WebsocketServer) receive(websocketConnection *WebsocketConnection) ([]byte, error) {
	websocketConnection.receiveMutex.Lock()
	defer websocketConnection.receiveMutex.Unlock()

	if event := server.onInfo(Event.NewInfo(
		Event.ReceivingMessage,
		"receiving message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:        Event.WebsocketConnection,
			Event.Address:     websocketConnection.GetIp(),
			Event.WebsocketId: websocketConnection.GetId(),
		}),
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	websocketConnection.websocketConnection.SetReadDeadline(time.Now().Add(time.Duration(server.config.ServerReadDeadlineMs) * time.Millisecond))
	_, messageBytes, err := websocketConnection.websocketConnection.ReadMessage()
	if err != nil {
		websocketConnection.Close()
		server.onWarning(Event.NewWarningNoOption(
			Event.NetworkError,
			err.Error(),
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:        Event.WebsocketConnection,
				Event.Address:     websocketConnection.GetIp(),
				Event.WebsocketId: websocketConnection.GetId(),
			}),
		))
		return nil, err
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.ReceivedMessage,
		"received message",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:        Event.WebsocketConnection,
			Event.Address:     websocketConnection.GetIp(),
			Event.WebsocketId: websocketConnection.GetId(),
		}),
	))
	return messageBytes, nil
}

// may only be called during the connections onConnectHandler.
func (server *WebsocketServer) Receive(websocketConnection *WebsocketConnection) ([]byte, error) {
	if websocketConnection.isAccepted {
		server.onWarning(Event.NewWarningNoOption(
			Event.ClientAlreadyAccepted,
			"websocketConnection is already accepted",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:        Event.WebsocketConnection,
				Event.Address:     websocketConnection.GetIp(),
				Event.WebsocketId: websocketConnection.GetId(),
			}),
		))
		return nil, errors.New("websocketConnection is already accepted")
	}
	return server.receive(websocketConnection)
}
