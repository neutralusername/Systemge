package WebsocketServer

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

type WebsocketClient struct {
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

// Disconnects the client and blocks until the client onDisconnectHandler has finished.
func (client *WebsocketClient) Close() error {
	client.closeMutex.Lock()
	defer client.closeMutex.Unlock()
	if client.isClosed {
		return errors.New("client is already closed")
	}
	client.isClosed = true
	client.websocketConnection.Close()
	if client.rateLimiterBytes != nil {
		client.rateLimiterBytes.Close()
	}
	if client.rateLimiterMsgs != nil {
		client.rateLimiterMsgs.Close()
	}
	close(client.stopChannel)
	return nil
}

// Returns the ip of the client.
func (client *WebsocketClient) GetIp() string {
	return client.websocketConnection.RemoteAddr().String()
}

// Returns the id of the client.
func (client *WebsocketClient) GetId() string {
	return client.id
}

// Sends a message to the client.
func (server *WebsocketServer) Send(client *WebsocketClient, messageBytes []byte) error {
	client.sendMutex.Lock()
	defer client.sendMutex.Unlock()

	if event := server.onInfo(Event.NewInfo(
		Event.SendingMessage,
		"sending message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:              Event.WebsocketConnection,
			Event.Address:           client.GetIp(),
			Event.TargetWebsocketId: client.GetId(),
			"bytes":                 string(messageBytes),
		}),
	)); !event.IsInfo() {
		server.failedSendCounter.Add(1)
		return event.GetError()
	}

	err := client.websocketConnection.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		server.failedSendCounter.Add(1)
		server.onWarning(Event.NewWarningNoOption(
			Event.NetworkError,
			err.Error(),
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:              Event.WebsocketConnection,
				Event.Address:           client.GetIp(),
				Event.TargetWebsocketId: client.GetId(),
				"bytes":                 string(messageBytes),
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
			Event.Address:           client.GetIp(),
			Event.TargetWebsocketId: client.GetId(),
			"bytes":                 string(messageBytes),
		}),
	))
	return nil
}

func (server *WebsocketServer) receive(client *WebsocketClient) ([]byte, error) {
	client.receiveMutex.Lock()
	defer client.receiveMutex.Unlock()

	if event := server.onInfo(Event.NewInfo(
		Event.ReceivingMessage,
		"receiving message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:        Event.WebsocketConnection,
			Event.Address:     client.GetIp(),
			Event.WebsocketId: client.GetId(),
		}),
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	client.websocketConnection.SetReadDeadline(time.Now().Add(time.Duration(server.config.ServerReadDeadlineMs) * time.Millisecond))
	_, messageBytes, err := client.websocketConnection.ReadMessage()
	if err != nil {
		client.Close()
		server.onWarning(Event.NewWarningNoOption(
			Event.NetworkError,
			err.Error(),
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:        Event.WebsocketConnection,
				Event.Address:     client.GetIp(),
				Event.WebsocketId: client.GetId(),
			}),
		))
		return nil, err
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.ReceivedMessage,
		"received message",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:        Event.WebsocketConnection,
			Event.Address:     client.GetIp(),
			Event.WebsocketId: client.GetId(),
		}),
	))
	return messageBytes, nil
}

// may only be called during the connections onConnectHandler.
func (server *WebsocketServer) Receive(client *WebsocketClient) ([]byte, error) {
	if client.isAccepted {
		server.onWarning(Event.NewWarningNoOption(
			Event.ClientAlreadyAccepted,
			"client is already accepted",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:        Event.WebsocketConnection,
				Event.Address:     client.GetIp(),
				Event.WebsocketId: client.GetId(),
			}),
		))
		return nil, errors.New("client is already accepted")
	}
	return server.receive(client)
}
