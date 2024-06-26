package Node

import (
	"Systemge/Utilities"
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

	// the minimum time which must pass between two messages from the websocketClient.
	// otherwise the message is ignored.
	messageCooldown time.Duration

	// the timestamp of the previous message from the websocketClient.
	// used to enforce messageCooldown.
	// updated automatically after every message to the current time.
	// can be set manually to a time in the future to block messages until that time.
	lastMessageTimestamp time.Time

	// if true, messages are handled as soon as they are received.
	// if false, messages are handled in the order they are received, one after the other.
	handleMessagesConcurrently bool
}

func newWebsocketClient(id string, websocketConn *websocket.Conn, onDisconnectHandler func(*WebsocketClient)) *WebsocketClient {
	websocketClient := &WebsocketClient{
		id:            id,
		websocketConn: websocketConn,
		stopChannel:   make(chan bool),

		messageCooldown:            DEFAULT_MESSAGE_COOLDOWN,
		lastMessageTimestamp:       time.Now(),
		handleMessagesConcurrently: DEFAULT_HANDLE_MESSAGES_CONCURRENTLY,
	}
	websocketClient.watchdog = time.AfterFunc(WATCHDOG_TIMEOUT, func() {
		watchdog := websocketClient.watchdog
		websocketClient.watchdog = nil
		watchdog.Stop()
		websocketConn.Close()
		onDisconnectHandler(websocketClient)
		close(websocketClient.stopChannel)
	})
	return websocketClient
}

func (websocketClient *WebsocketClient) GetLastMessageTimestamp() time.Time {
	return websocketClient.lastMessageTimestamp
}

func (websocketClient *WebsocketClient) SetLastMessageTimestamp(lastMessageTimestamp time.Time) {
	websocketClient.lastMessageTimestamp = lastMessageTimestamp
}

func (websocketClient *WebsocketClient) SetHandleMessagesConcurrently(handleMessagesConcurrently bool) {
	websocketClient.handleMessagesConcurrently = handleMessagesConcurrently
}

func (websocketClient *WebsocketClient) GetHandleMessagesConcurrently() bool {
	return websocketClient.handleMessagesConcurrently
}

func (websocketClient *WebsocketClient) SetMessageCooldown(messageCooldown time.Duration) {
	websocketClient.messageCooldown = messageCooldown
}

func (websocketClient *WebsocketClient) GetMessageCooldown() time.Duration {
	return websocketClient.messageCooldown
}

// Resets the watchdog timer to its initial value
func (websocketClient *WebsocketClient) ResetWatchdog() {
	websocketClient.watchdogMutex.Lock()
	defer websocketClient.watchdogMutex.Unlock()
	if websocketClient.watchdog == nil {
		return
	}
	websocketClient.watchdog.Reset(WATCHDOG_TIMEOUT)
}

// Disconnects the websocketClient and blocks until after the onDisconnectHandler is called
func (websocketClient *WebsocketClient) Disconnect() {
	websocketClient.watchdogMutex.Lock()
	defer websocketClient.watchdogMutex.Unlock()
	if websocketClient.watchdog == nil {
		return
	}
	websocketClient.watchdog.Reset(0)
	<-websocketClient.stopChannel
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
	return messageBytes, err
}

func (node *Node) addWebsocketConn(websocketConn *websocket.Conn) *WebsocketClient {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	websocketId := "#" + node.randomizer.GenerateRandomString(16, Utilities.ALPHA_NUMERIC)
	for _, exists := node.websocketClients[websocketId]; exists; {
		websocketId = "#" + node.randomizer.GenerateRandomString(16, Utilities.ALPHA_NUMERIC)
	}
	websocketClient := newWebsocketClient(websocketId, websocketConn, func(websocketClient *WebsocketClient) {
		node.websocketComponent.OnDisconnectHandler(node, websocketClient)
		node.removeWebsocketClient(websocketClient)
	})
	node.websocketClients[websocketId] = websocketClient
	node.websocketClientGroups[websocketId] = make(map[string]bool)
	return websocketClient
}

func (node *Node) removeWebsocketClient(websocketClient *WebsocketClient) {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	delete(node.websocketClients, websocketClient.GetId())
	for groupId := range node.websocketClientGroups[websocketClient.GetId()] {
		delete(node.websocketClientGroups[websocketClient.GetId()], groupId)
		delete(node.WebsocketGroups[groupId], websocketClient.GetId())
		if len(node.WebsocketGroups[groupId]) == 0 {
			delete(node.WebsocketGroups, groupId)
		}
	}
}

func (node *Node) WebsocketClientExists(websocketId string) bool {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	_, exists := node.websocketClients[websocketId]
	return exists
}
