package MessageBrokerServer

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/TCP"
	"net"
	"sync"
	"time"
)

type Client struct {
	name         string
	netConn      net.Conn
	messageQueue chan *Message.Message
	watchdog     *time.Timer

	subscribedTopics   map[string]bool
	openSyncRequests   map[string]*Message.Message
	deliverImmediately bool

	mutex        sync.Mutex
	sendMutex    sync.Mutex
	receiveMutex sync.Mutex
}

func NewClienn(name string, netConn net.Conn) *Client {
	return &Client{
		name:               name,
		netConn:            netConn,
		messageQueue:       make(chan *Message.Message, CLIENT_MESSAGE_QUEUE_SIZE),
		watchdog:           nil,
		openSyncRequests:   map[string]*Message.Message{},
		subscribedTopics:   map[string]bool{},
		deliverImmediately: DELIVER_IMMEDIATELY_DEFAULT,
	}
}

func (client *Client) send(message *Message.Message) error {
	client.sendMutex.Lock()
	defer client.sendMutex.Unlock()
	return TCP.Send(client.netConn, message.Serialize(), DEFAULT_TCP_TIMEOUT)
}

func (client *Client) receive() ([]byte, error) {
	client.receiveMutex.Lock()
	defer client.receiveMutex.Unlock()
	messageBytes, err := TCP.Receive(client.netConn, 0)
	if err != nil {
		return nil, err
	}
	return messageBytes, nil
}

func (client *Client) resetWatchdog() error {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	if client.watchdog == nil {
		return Error.New("Watchdog is not set for client \""+client.name+"\"", nil)
	}
	client.watchdog.Reset((WATCHDOG_TIMEOUT))
	return nil
}

func (client *Client) disconnect() error {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	if client.watchdog == nil {
		return Error.New("Watchdog is not set for client \""+client.name+"\"", nil)
	}
	client.watchdog.Reset(0)
	client.watchdog = nil
	client.netConn.Close()
	return nil
}

func (server *Server) addClient(client *Client) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.clients[client.name] != nil {
		return Error.New("client with name \""+client.name+"\" already exists", nil)
	}
	server.clients[client.name] = client
	client.watchdog = time.AfterFunc(WATCHDOG_TIMEOUT, func() {
		server.removeClient(client)
	})
	return nil
}

func (server *Server) removeClient(client *Client) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	if server.clients[client.name] == nil {
		return Error.New("Subscriber with name \""+client.name+"\" does not exist", nil)
	}
	client.watchdog = nil
	client.netConn.Close()
	for messageType := range client.subscribedTopics {
		delete(server.subscriptions[messageType], client.name)
	}
	for _, message := range client.openSyncRequests {
		delete(server.openSyncRequests, message.GetSyncRequestToken())
	}
	delete(server.clients, client.name)
	return nil
}
