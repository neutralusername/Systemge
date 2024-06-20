package WebsocketClient

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	id            string
	websocketConn *websocket.Conn

	watchdogMutex sync.Mutex
	receiveMutex  sync.Mutex
	sendMutex     sync.Mutex
	stopChannel   chan bool

	// the watchdog timer is reset every time a message is received.
	// if the timer expires, the client is disconnected.
	// if the timer is nil, the client is already disconnected.
	watchdog *time.Timer

	// the minimum time which must pass between two messages from the client.
	// otherwise the message is ignored.
	messageCooldown time.Duration

	// the timestamp of the previous message from the client.
	// used to enforce messageCooldown.
	// updated automatically after every message to the current time.
	// can be set manually to a time in the future to block messages until that time.
	lastMessageTimestamp time.Time

	// if true, messages are handled as soon as they are received.
	// if false, messages are handled in the order they are received, one after the other.
	handleMessagesConcurrently bool
}

func New(id string, websocketConn *websocket.Conn, onDisconnectHandler func(*Client)) *Client {
	client := &Client{
		id:            id,
		websocketConn: websocketConn,
		stopChannel:   make(chan bool),

		messageCooldown:            DEFAULT_MESSAGE_COOLDOWN,
		lastMessageTimestamp:       time.Now(),
		handleMessagesConcurrently: DEFAULT_HANDLE_MESSAGES_CONCURRENTLY,
	}
	client.watchdog = time.AfterFunc(WATCHDOG_TIMEOUT, func() {
		watchdog := client.watchdog
		client.watchdog = nil
		watchdog.Stop()
		websocketConn.Close()
		onDisconnectHandler(client)
		close(client.stopChannel)
	})
	return client
}

func (client *Client) GetLastMessageTimestamp() time.Time {
	return client.lastMessageTimestamp
}

func (client *Client) SetLastMessageTimestamp(lastMessageTimestamp time.Time) {
	client.lastMessageTimestamp = lastMessageTimestamp
}

func (client *Client) SetHandleMessagesConcurrently(handleMessagesConcurrently bool) {
	client.handleMessagesConcurrently = handleMessagesConcurrently
}

func (client *Client) GetHandleMessagesConcurrently() bool {
	return client.handleMessagesConcurrently
}

func (client *Client) SetMessageCooldown(messageCooldown time.Duration) {
	client.messageCooldown = messageCooldown
}

func (client *Client) GetMessageCooldown() time.Duration {
	return client.messageCooldown
}

// Resets the watchdog timer to its initial value
func (client *Client) ResetWatchdog() {
	client.watchdogMutex.Lock()
	defer client.watchdogMutex.Unlock()
	if client.watchdog == nil {
		return
	}
	client.watchdog.Reset(WATCHDOG_TIMEOUT)
}

// Disconnects the client and blocks until after the onDisconnectHandler is called
func (client *Client) Disconnect() {
	client.watchdogMutex.Lock()
	defer client.watchdogMutex.Unlock()
	if client.watchdog == nil {
		return
	}
	client.watchdog.Reset(0)
	<-client.stopChannel
}

func (client *Client) GetIp() string {
	return client.websocketConn.RemoteAddr().String()
}

func (client *Client) GetId() string {
	return client.id
}

func (client *Client) Send(messageBytes []byte) error {
	client.sendMutex.Lock()
	defer client.sendMutex.Unlock()
	return client.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
}

func (client *Client) Receive() ([]byte, error) {
	client.receiveMutex.Lock()
	defer client.receiveMutex.Unlock()
	_, messageBytes, err := client.websocketConn.ReadMessage()
	return messageBytes, err
}
