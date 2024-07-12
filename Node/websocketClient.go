package Node

import (
	"Systemge/Error"
	"Systemge/Message"
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

	// the timestamp of the previous message from the websocketClient.
	// used to enforce messageCooldown.
	// updated automatically after every message to the current time.
	// can be set manually to a time in the future to block messages until that time.
	lastMessageTimestamp time.Time
}

func (node *Node) newWebsocketClient(id string, websocketConn *websocket.Conn, onDisconnectHandler func(*WebsocketClient)) *WebsocketClient {
	websocketClient := &WebsocketClient{
		id:            id,
		websocketConn: websocketConn,
		stopChannel:   make(chan bool),

		lastMessageTimestamp: time.Now(),
	}

	websocketClient.watchdogMutex.Lock()
	//without the mutex's there is a rare racecondition in which the anonymous function is called (if duration is extremely short or 0) before websocketClient.watchdog is assigned its value which results in a panic
	websocketClient.watchdog = time.AfterFunc(time.Duration(node.websocketComponent.GetWebsocketComponentConfig().ClientWatchdogTimeoutMs)*time.Millisecond, func() {
		websocketClient.watchdogMutex.Lock()
		defer websocketClient.watchdogMutex.Unlock()
		if websocketClient.watchdog == nil {
			return
		}
		websocketClient.watchdog.Stop()
		websocketClient.watchdog = nil
		websocketConn.Close()
		onDisconnectHandler(websocketClient)
		close(websocketClient.stopChannel)
	})
	websocketClient.watchdogMutex.Unlock()
	return websocketClient
}

func (websocketClient *WebsocketClient) GetLastMessageTimestamp() time.Time {
	return websocketClient.lastMessageTimestamp
}

func (websocketClient *WebsocketClient) SetLastMessageTimestamp(lastMessageTimestamp time.Time) {
	websocketClient.lastMessageTimestamp = lastMessageTimestamp
}

// Resets the watchdog timer to its initial value
func (node *Node) ResetWatchdog(websocketClient *WebsocketClient) {
	websocketClient.watchdogMutex.Lock()
	defer websocketClient.watchdogMutex.Unlock()
	if websocketClient.watchdog == nil {
		return
	}
	websocketClient.watchdog.Reset(time.Duration(node.websocketComponent.GetWebsocketComponentConfig().ClientWatchdogTimeoutMs) * time.Millisecond)
}

// Disconnects the websocketClient and blocks until after the onDisconnectHandler is called
func (websocketClient *WebsocketClient) Disconnect() {
	websocketClient.watchdogMutex.Lock()
	if websocketClient.watchdog == nil {
		websocketClient.watchdogMutex.Unlock()
		return
	}
	websocketClient.watchdog.Reset(0)
	websocketClient.watchdogMutex.Unlock()
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

func (websocketClient *WebsocketClient) Receive() (*Message.Message, error) {
	websocketClient.receiveMutex.Lock()
	defer websocketClient.receiveMutex.Unlock()
	_, messageBytes, err := websocketClient.websocketConn.ReadMessage()
	if err != nil {
		return nil, Error.New("failed to receive message", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil {
		return nil, Error.New("failed to deserialize message \""+string(messageBytes)+"\"", nil)
	}
	message = Message.NewAsync(message.GetTopic(), websocketClient.GetId(), message.GetPayload())
	return message, err
}

func (node *Node) addWebsocketConn(websocketConn *websocket.Conn) *WebsocketClient {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	websocketId := "#" + node.randomizer.GenerateRandomString(16, Utilities.ALPHA_NUMERIC)
	for _, exists := node.websocketClients[websocketId]; exists; {
		websocketId = "#" + node.randomizer.GenerateRandomString(16, Utilities.ALPHA_NUMERIC)
	}
	websocketClient := node.newWebsocketClient(websocketId, websocketConn, func(websocketClient *WebsocketClient) {
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
