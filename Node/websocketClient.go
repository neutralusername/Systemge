package Node

import (
	"Systemge/Error"
	"sync"
	"time"

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

	expired      bool
	disconnected bool

	// the timestamp of the previous message from the websocketClient.
	// used to enforce messageCooldown.
	// updated automatically after every message to the current time.
	// can be set manually to a time in the future to block messages until that time.
	lastMessageTimestamp time.Time
}

func (websocket *websocketComponent) newWebsocketClient(id string, websocketConn *websocket.Conn) *WebsocketClient {
	websocketClient := &WebsocketClient{
		id:            id,
		websocketConn: websocketConn,
		stopChannel:   make(chan bool),

		lastMessageTimestamp: time.Now(),
	}

	websocketClient.watchdogMutex.Lock()
	defer websocketClient.watchdogMutex.Unlock()
	websocketClient.watchdog = time.AfterFunc(time.Duration(websocket.application.GetWebsocketComponentConfig().ClientWatchdogTimeoutMs)*time.Millisecond, func() {
		websocketClient.expired = true
		websocketClient.watchdogMutex.Lock()
		defer websocketClient.watchdogMutex.Unlock()
		if websocketClient.watchdog == nil || (!websocketClient.disconnected && !websocketClient.expired) {
			return
		}
		websocketClient.watchdog.Stop()
		websocketClient.watchdog = nil
		websocketConn.Close()
		websocket.onDisconnectWraper(websocketClient)
		websocket.removeWebsocketClient(websocketClient)
		close(websocketClient.stopChannel)
	})
	return websocketClient
}

// Resets the watchdog timer to its initial value
func (node *Node) ResetWatchdog(websocketClient *WebsocketClient) error {
	if websocketClient == nil {
		return Error.New("websocketClient is nil", nil)
	}
	websocketClient.watchdogMutex.Lock()
	defer websocketClient.watchdogMutex.Unlock()
	if websocketClient.watchdog == nil || websocketClient.disconnected {
		return Error.New("websocketClient is disconnected", nil)
	}
	websocketClient.expired = false
	websocketClient.watchdog.Reset(time.Duration(node.websocket.application.GetWebsocketComponentConfig().ClientWatchdogTimeoutMs) * time.Millisecond)
	return nil
}

// Disconnects the websocketClient and blocks until the websocketClient is disconnected.
func (websocketClient *WebsocketClient) Disconnect() error {
	if websocketClient == nil {
		return Error.New("websocketClient is nil", nil)
	}
	websocketClient.watchdogMutex.Lock()
	if websocketClient.watchdog == nil || websocketClient.disconnected {
		websocketClient.watchdogMutex.Unlock()
		return Error.New("websocketClient is already disconnected", nil)
	}
	websocketClient.disconnected = true
	websocketClient.watchdog.Reset(0)
	websocketClient.watchdogMutex.Unlock()
	<-websocketClient.stopChannel
	return nil
}

func (websocketClient *WebsocketClient) GetLastMessageTimestamp() time.Time {
	return websocketClient.lastMessageTimestamp
}

func (websocketClient *WebsocketClient) SetLastMessageTimestamp(lastMessageTimestamp time.Time) {
	websocketClient.lastMessageTimestamp = lastMessageTimestamp
}

func (websocketClient *WebsocketClient) GetIp() string {
	return websocketClient.websocketConn.RemoteAddr().String()
}

func (websocketClient *WebsocketClient) GetId() string {
	return websocketClient.id
}

func (websocketClient *WebsocketClient) Send(messageBytes []byte) error {
	websocketClient.sendMutex.Lock()
	defer websocketClient.sendMutex.Unlock()
	return websocketClient.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
}

func (websocketClient *WebsocketClient) Receive() ([]byte, error) {
	websocketClient.receiveMutex.Lock()
	defer websocketClient.receiveMutex.Unlock()
	_, messageBytes, err := websocketClient.websocketConn.ReadMessage()
	if err != nil {
		return nil, Error.New("failed to receive message", err)
	}
	return messageBytes, err
}
