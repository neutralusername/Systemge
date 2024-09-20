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

	pastOnConnectHandler bool

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
func (client *WebsocketClient) GetIp() string { rename to address when i can
	return client.websocketConnection.RemoteAddr().String()
}

// Returns the id of the client.
func (client *WebsocketClient) GetId() string {
	return client.id
}

// Sends a message to the client.
func (client *WebsocketClient) Send(messageBytes []byte) error {
	client.sendMutex.Lock()
	defer client.sendMutex.Unlock()
	return client.websocketConnection.WriteMessage(websocket.TextMessage, messageBytes)
}

func (client *WebsocketClient) receive() ([]byte, *Event.Event) {
	client.receiveMutex.Lock()
	defer client.receiveMutex.Unlock()
	_, messageBytes, err := client.websocketConnection.ReadMessage()
	if err != nil {
		return nil, Event.New("failed to receive message", err)
	}
	return messageBytes, err
}

// may only be called during the connections onConnectHandler.
func (client *WebsocketClient) Receive() ([]byte, error) {
	if client.pastOnConnectHandler {
		return nil, Event.New("may only be called during the connections onConnectHandler", nil)
	}
	return client.receive()
}
