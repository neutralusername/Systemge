package Node

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

	rateLimiterBytes *Tools.RateLimiter
	rateLimiterMsgs  *Tools.RateLimiter

	expired      bool
	disconnected bool
}

func (websocket *websocketComponent) newWebsocketClient(id string, websocketConn *websocket.Conn) *WebsocketClient {
	websocketClient := &WebsocketClient{
		id:            id,
		websocketConn: websocketConn,
		stopChannel:   make(chan bool),
	}
	if websocket.config.ClientRateLimiterBytes != nil {
		websocketClient.rateLimiterBytes = Tools.NewRateLimiter(websocket.config.ClientRateLimiterBytes)
	}
	if websocket.config.ClientRateLimiterMsgs != nil {
		websocketClient.rateLimiterMsgs = Tools.NewRateLimiter(websocket.config.ClientRateLimiterMsgs)
	}
	websocketClient.websocketConn.SetReadLimit(int64(websocket.config.IncomingMessageByteLimit))

	websocketClient.watchdogMutex.Lock()
	defer websocketClient.watchdogMutex.Unlock()
	websocketClient.watchdog = time.AfterFunc(time.Duration(websocket.config.ClientWatchdogTimeoutMs)*time.Millisecond, func() {
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
		websocket.onDisconnectWraper(websocketClient)
		websocket.removeWebsocketClient(websocketClient)
		close(websocketClient.stopChannel)
	})
	return websocketClient
}

// Resets the watchdog timer to its initial value
func (node *Node) ResetWatchdog(websocketClient *WebsocketClient) error {
	if websocket := node.websocket; websocket != nil {
		if websocketClient == nil {
			return Error.New("websocketClient is nil", nil)
		}
		websocketClient.watchdogMutex.Lock()
		defer websocketClient.watchdogMutex.Unlock()
		if websocketClient.watchdog == nil || websocketClient.disconnected {
			return Error.New("websocketClient is disconnected", nil)
		}
		websocketClient.expired = false
		websocketClient.watchdog.Reset(time.Duration(websocket.config.ClientWatchdogTimeoutMs) * time.Millisecond)
		return nil
	}
	return Error.New("websocket is nil", nil)
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
