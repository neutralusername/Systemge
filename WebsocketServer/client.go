package WebsocketServer

import (
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

type WebsocketClient struct {
	id            string
	websocketConn *websocket.Conn

	watchdogMutex sync.Mutex
	receiveMutex  sync.Mutex
	sendMutex     sync.Mutex
	stopChannel   chan bool

	// the watchdog timer is reset every time a message is received.
	// if the timer expires, the websocketClient is disconnected.
	// if the timer is nil, the websocketClient is already disconnected.
	watchdog *time.Timer

	rateLimiterBytes *Tools.TokenBucketRateLimiter
	rateLimiterMsgs  *Tools.TokenBucketRateLimiter

	expired      bool
	disconnected bool
}

func (server *WebsocketServer) newWebsocketClient(id string, websocketConn *websocket.Conn) *WebsocketClient {
	websocketClient := &WebsocketClient{
		id:            id,
		websocketConn: websocketConn,
		stopChannel:   make(chan bool),
	}
	if server.config.ClientRateLimiterBytes != nil {
		websocketClient.rateLimiterBytes = Tools.NewTokenBucketRateLimiter(server.config.ClientRateLimiterBytes)
	}
	if server.config.ClientRateLimiterMessages != nil {
		websocketClient.rateLimiterMsgs = Tools.NewTokenBucketRateLimiter(server.config.ClientRateLimiterMessages)
	}
	websocketClient.websocketConn.SetReadLimit(int64(server.config.IncomingMessageByteLimit))

	websocketClient.watchdogMutex.Lock()
	defer websocketClient.watchdogMutex.Unlock()
	websocketClient.watchdog = time.AfterFunc(time.Duration(server.config.ClientWatchdogTimeoutMs)*time.Millisecond, func() {
		websocketClient.expired = true
		websocketClient.watchdogMutex.Lock()
		defer websocketClient.watchdogMutex.Unlock()
		if websocketClient.watchdog == nil || (!websocketClient.disconnected && !websocketClient.expired) {
			return
		}
		websocketClient.watchdog.Stop()
		websocketClient.watchdog = nil
		websocketConn.Close()
		if websocketClient.rateLimiterBytes != nil {
			websocketClient.rateLimiterBytes.Stop()
		}
		if websocketClient.rateLimiterMsgs != nil {
			websocketClient.rateLimiterMsgs.Stop()
		}
		if server.onDisconnectHandler != nil {
			server.onDisconnectHandler(websocketClient)
		}
		server.removeWebsocketClient(websocketClient)
		close(websocketClient.stopChannel)
	})
	return websocketClient
}

// Resets the watchdog timer to its initial value.
func (server *WebsocketServer) ResetWatchdog(websocketClient *WebsocketClient) error {
	return server.resetWatchdog(websocketClient)
}

func (server *WebsocketServer) resetWatchdog(websocketClient *WebsocketClient) error {
	if websocketClient == nil {
		return Error.New("websocketClient is nil", nil)
	}
	websocketClient.watchdogMutex.Lock()
	defer websocketClient.watchdogMutex.Unlock()
	if websocketClient.watchdog == nil || websocketClient.disconnected {
		return Error.New("websocketClient is disconnected", nil)
	}
	websocketClient.expired = false
	websocketClient.watchdog.Reset(time.Duration(server.config.ClientWatchdogTimeoutMs) * time.Millisecond)
	return nil
}

// Disconnects the websocketClient and blocks until the websocketClients onDisconnectHandler has finished.
func (client *WebsocketClient) Disconnect() error {
	if client == nil {
		return Error.New("websocketClient is nil", nil)
	}
	client.watchdogMutex.Lock()
	if client.watchdog == nil || client.disconnected {
		client.watchdogMutex.Unlock()
		return Error.New("websocketClient is already disconnected", nil)
	}
	client.disconnected = true
	client.watchdog.Reset(0)
	client.watchdogMutex.Unlock()
	<-client.stopChannel
	return nil
}

// Returns the ip of the websocketClient.
func (client *WebsocketClient) GetIp() string {
	return client.websocketConn.RemoteAddr().String()
}

// Returns the id of the websocketClient.
func (client *WebsocketClient) GetId() string {
	return client.id
}

// Sends a message to the websocketClient.
func (client *WebsocketClient) Send(messageBytes []byte) error {
	client.sendMutex.Lock()
	defer client.sendMutex.Unlock()
	return client.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
}

func (client *WebsocketClient) receive() ([]byte, error) {
	client.receiveMutex.Lock()
	defer client.receiveMutex.Unlock()
	_, messageBytes, err := client.websocketConn.ReadMessage()
	if err != nil {
		return nil, Error.New("failed to receive message", err)
	}
	return messageBytes, err
}
