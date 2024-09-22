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
	if event := server.onInfo(Event.New(
		Event.SendingMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":              "writing to websocket connection",
			"type":              "websocket",
			"address":           client.GetIp(),
			"targetWebsocketId": client.GetId(),
			"bytes":             string(messageBytes),
		}),
	)); event.IsError() {
		server.failedSendCounter.Add(1)
		return event
	}
	err := client.websocketConnection.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		server.failedSendCounter.Add(1)
		return server.onError(Event.New(
			Event.NetworkError,
			server.GetServerContext().Merge(Event.Context{
				"error":             "failed to write to websocket connection",
				"type":              "websocket",
				"address":           client.GetIp(),
				"targetWebsocketId": client.GetId(),
				"bytes":             string(messageBytes),
			}),
		))
	}
	server.outgoigMessageCounter.Add(1)
	server.bytesSentCounter.Add(uint64(len(messageBytes)))
	return server.onInfo(Event.New(
		Event.SentMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":              "wrote to websocket connection",
			"type":              "websocket",
			"address":           client.GetIp(),
			"targetWebsocketId": client.GetId(),
			"bytes":             string(messageBytes),
		}),
	))
}

func (server *WebsocketServer) receive(client *WebsocketClient) ([]byte, *Event.Event) {
	client.receiveMutex.Lock()
	defer client.receiveMutex.Unlock()
	if event := server.onInfo(Event.New(
		Event.ReceivingMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":        "receiving message from client",
			"type":        "websocket",
			"address":     client.GetIp(),
			"websocketId": client.GetId(),
		}),
	)); event.IsError() {
		return nil, event
	}
	client.websocketConnection.SetReadDeadline(time.Now().Add(time.Duration(server.config.ServerReadDeadlineMs) * time.Millisecond))
	_, messageBytes, err := client.websocketConnection.ReadMessage()
	if err != nil {
		client.Close()
		event := server.onError(Event.New(
			Event.NetworkError,
			server.GetServerContext().Merge(Event.Context{
				"error":       "failed to receive message from client",
				"type":        "websocket",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		))
		return nil, event
	}
	return messageBytes, server.onInfo(Event.New(
		Event.ReceivedMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":        "received message from client",
			"type":        "websocket",
			"address":     client.GetIp(),
			"websocketId": client.GetId(),
		}),
	))
}

// may only be called during the connections onConnectHandler.
func (server *WebsocketServer) Receive(client *WebsocketClient) ([]byte, error) {
	if client.isAccepted {
		return nil, server.onError(Event.New(
			Event.ClientAlreadyAccepted,
			server.GetServerContext().Merge(Event.Context{
				"error":       "client is already accepted",
				"type":        "websocket",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		))
	}
	return server.receive(client)
}
