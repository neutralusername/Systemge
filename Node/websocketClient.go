package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tools"
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

func (node *Node) newWebsocketClient(id string, websocketConn *websocket.Conn) *WebsocketClient {
	websocketClient := &WebsocketClient{
		id:            id,
		websocketConn: websocketConn,
		stopChannel:   make(chan bool),

		lastMessageTimestamp: time.Now(),
	}

	websocketClient.watchdogMutex.Lock()
	defer websocketClient.watchdogMutex.Unlock()
	websocketClient.watchdog = time.AfterFunc(time.Duration(node.GetWebsocketComponent().GetWebsocketComponentConfig().ClientWatchdogTimeoutMs)*time.Millisecond, func() {
		websocketClient.expired = true
		websocketClient.watchdogMutex.Lock()
		defer websocketClient.watchdogMutex.Unlock()
		if websocketClient.watchdog == nil || (!websocketClient.disconnected && !websocketClient.expired) {
			return
		}
		websocketClient.watchdog.Stop()
		websocketClient.watchdog = nil
		websocketConn.Close()
		node.GetWebsocketComponent().OnDisconnectHandler(node, websocketClient)
		node.removeWebsocketClient(websocketClient)
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
	websocketClient.watchdog.Reset(time.Duration(node.GetWebsocketComponent().GetWebsocketComponentConfig().ClientWatchdogTimeoutMs) * time.Millisecond)
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
	websocketId := "#" + node.randomizer.GenerateRandomString(16, Tools.ALPHA_NUMERIC)
	for _, exists := node.websocketClients[websocketId]; exists; {
		websocketId = "#" + node.randomizer.GenerateRandomString(16, Tools.ALPHA_NUMERIC)
	}
	websocketClient := node.newWebsocketClient(websocketId, websocketConn)
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
		delete(node.websocketGroups[groupId], websocketClient.GetId())
		if len(node.websocketGroups[groupId]) == 0 {
			delete(node.websocketGroups, groupId)
		}
	}
}

func (node *Node) WebsocketClientExists(websocketId string) bool {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	_, exists := node.websocketClients[websocketId]
	return exists
}
