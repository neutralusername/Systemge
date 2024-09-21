package WebsocketServer

import (
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

	watchdogMutex sync.Mutex
	receiveMutex  sync.Mutex
	sendMutex     sync.Mutex
	stopChannel   chan bool

	// the watchdog timer is reset every time a message is received.
	// if the timer expires, the client is disconnected.
	// if the timer is nil, the client is already disconnected.
	watchdog *time.Timer

	rateLimiterBytes *Tools.TokenBucketRateLimiter
	rateLimiterMsgs  *Tools.TokenBucketRateLimiter

	expired      bool
	disconnected bool
}

func (server *WebsocketServer) newClient(id string, websocketConnection *websocket.Conn) *WebsocketClient {
	client := &WebsocketClient{
		id:                  id,
		websocketConnection: websocketConnection,
		stopChannel:         make(chan bool),
	}
	if server.config.ClientRateLimiterBytes != nil {
		client.rateLimiterBytes = Tools.NewTokenBucketRateLimiter(server.config.ClientRateLimiterBytes)
	}
	if server.config.ClientRateLimiterMessages != nil {
		client.rateLimiterMsgs = Tools.NewTokenBucketRateLimiter(server.config.ClientRateLimiterMessages)
	}
	client.websocketConnection.SetReadLimit(int64(server.config.IncomingMessageByteLimit))

	client.watchdogMutex.Lock()
	defer client.watchdogMutex.Unlock()
	client.watchdog = time.AfterFunc(time.Duration(server.config.ClientWatchdogTimeoutMs)*time.Millisecond, func() {
		client.expired = true
		client.watchdogMutex.Lock()
		defer client.watchdogMutex.Unlock()
		if client.watchdog == nil || (!client.disconnected && !client.expired) {
			return
		}
		client.watchdog.Stop()
		client.watchdog = nil
		websocketConnection.Close()
		if client.rateLimiterBytes != nil {
			client.rateLimiterBytes.Close()
		}
		if client.rateLimiterMsgs != nil {
			client.rateLimiterMsgs.Close()
		}
		if server.onDisconnectHandler != nil {
			server.onDisconnectHandler(client)
		}
		server.removeClient(client)
		close(client.stopChannel)
	})
	return client
}

// Resets the watchdog timer to its initial value.
func (server *WebsocketServer) ResetWatchdog(client *WebsocketClient) error {
	if client == nil {
		return Event.New("client is nil", nil)
	}
	client.watchdogMutex.Lock()
	defer client.watchdogMutex.Unlock()
	if client.watchdog == nil || client.disconnected {
		return Event.New("client is disconnected", nil)
	}
	client.expired = false
	client.watchdog.Reset(time.Duration(server.config.ClientWatchdogTimeoutMs) * time.Millisecond)
	return nil
}

// Disconnects the client and blocks until the client onDisconnectHandler has finished.
func (client *WebsocketClient) Disconnect() error {
	if client == nil {
		return Event.New("client is nil", nil)
	}
	client.watchdogMutex.Lock()
	if client.watchdog == nil || client.disconnected {
		client.watchdogMutex.Unlock()
		return Event.New("client is already disconnected", nil)
	}
	client.disconnected = true
	client.watchdog.Reset(0)
	client.watchdogMutex.Unlock()
	<-client.stopChannel
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
func (server *WebsocketServer) Send(client *WebsocketClient, messageBytes []byte) *Event.Event {
	client.sendMutex.Lock()
	defer client.sendMutex.Unlock()
	if event := server.onInfo(Event.New(
		Event.SendingMessage,
		server.GetServerContext().Merge(Event.Context{
			"messageType":       "websocketWrite",
			"bytes":             string(messageBytes),
			"info":              "writing to websocket connection",
			"targetWebsocketId": client.GetId(),
		}),
	)); event.IsError() {
		server.failedMessageCounter.Add(1)
		return event
	}
	err := client.websocketConnection.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		server.failedMessageCounter.Add(1)
		return server.onError(Event.New(
			Event.FailedToSendMessage,
			server.GetServerContext().Merge(Event.Context{
				"messageType":       "websocketWrite",
				"bytes":             string(messageBytes),
				"error":             "failed to write to websocket connection",
				"targetWebsocketId": client.GetId(),
			}),
		))
	}
	server.outgoigMessageCounter.Add(1)
	server.bytesSentCounter.Add(uint64(len(messageBytes)))
	return server.onInfo(Event.New(
		Event.SentMessage,
		server.GetServerContext().Merge(Event.Context{
			"messageType":       "websocketWrite",
			"bytes":             string(messageBytes),
			"info":              "wrote to websocket connection",
			"targetWebsocketId": client.GetId(),
		}),
	))
}

func (server *WebsocketServer) receive(client *WebsocketClient) ([]byte, *Event.Event) {
	client.receiveMutex.Lock()
	defer client.receiveMutex.Unlock()
	if event := server.onInfo(Event.New(
		Event.ReceivingMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":               "receiving message from client",
			"serviceRoutineType": "handleMessages",
			"address":            client.GetIp(),
			"websocketId":        client.GetId(),
		}),
	)); event.IsError() {
		return nil, event
	}
	_, messageBytes, err := client.websocketConnection.ReadMessage()
	if err != nil {
		event := server.onError(Event.New(
			Event.FailedToReceiveMessage,
			server.GetServerContext().Merge(Event.Context{
				"error":              "failed to receive message from client",
				"serviceRoutineType": "handleMessages",
				"address":            client.GetIp(),
				"websocketId":        client.GetId(),
			}),
		))
		return nil, event
	}
	return messageBytes, server.onInfo(Event.New(
		Event.ReceivedMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":               "received message from client",
			"serviceRoutineType": "handleMessages",
			"address":            client.GetIp(),
			"websocketId":        client.GetId(),
		}),
	))
}

// may only be called during the connections onConnectHandler.
func (client *WebsocketClient) Receive() ([]byte, error) {
	if client.isAccepted {
		return nil, Event.New("may only be called during the connections onConnectHandler", nil)
	}
	return client.receive()
}
