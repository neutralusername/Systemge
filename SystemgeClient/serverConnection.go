package SystemgeClient

import (
	"net"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

// serverConnections are used to propagate async and sync requests and receive sync responses for their corresponding requests
type serverConnection struct {
	netConn        net.Conn
	endpointConfig *Config.TcpEndpoint
	name           string
	sendMutex      sync.Mutex
	topics         map[string]bool
	topicsMutex    sync.Mutex
	isTransient    bool

	rateLimiterBytes *Tools.TokenBucketRateLimiter
	rateLimiterMsgs  *Tools.TokenBucketRateLimiter
	tcpBuffer        []byte

	stopChannel chan bool //closing of this channel indicates that the server connection has finished its ongoing tasks.
}

func (client *SystemgeClient) newServerConnection(netConn net.Conn, endpoint *Config.TcpEndpoint, name string, topics map[string]bool) *serverConnection {
	serverConnection := &serverConnection{
		netConn:        netConn,
		endpointConfig: endpoint,
		name:           name,
		topics:         topics,
		stopChannel:    make(chan bool),
	}
	if client.config.RateLimiterBytes != nil {
		serverConnection.rateLimiterBytes = Tools.NewTokenBucketRateLimiter(client.config.RateLimiterBytes)
	}
	if client.config.RateLimiterMessages != nil {
		serverConnection.rateLimiterMsgs = Tools.NewTokenBucketRateLimiter(client.config.RateLimiterMessages)
	}
	return serverConnection
}

func (serverConnection *serverConnection) receiveMessage(bufferSize uint32, incomingMessageByteLimit uint64) ([]byte, error) {
	completedMsgBytes := []byte{}
	for {
		if incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > incomingMessageByteLimit {
			return nil, Error.New("Incoming message byte limit exceeded", nil)
		}
		for i, b := range serverConnection.tcpBuffer {
			if b == Tcp.ENDOFMESSAGE {
				completedMsgBytes = append(completedMsgBytes, serverConnection.tcpBuffer[:i]...)
				serverConnection.tcpBuffer = serverConnection.tcpBuffer[i+1:]
				if incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > incomingMessageByteLimit {
					return nil, Error.New("Incoming message byte limit exceeded", nil)
				}
				return completedMsgBytes, nil
			}
		}
		completedMsgBytes = append(completedMsgBytes, serverConnection.tcpBuffer...)
		receivedMessageBytes, _, err := Tcp.Receive(serverConnection.netConn, 0, bufferSize)
		if err != nil {
			return nil, Error.New("Failed to refill tcp buffer", err)
		}
		serverConnection.tcpBuffer = receivedMessageBytes
	}
}

// async messages and sync requests are sent to server connections
func (client *SystemgeClient) messageServerConnection(serverConnection *serverConnection, message *Message.Message) error {
	serverConnection.sendMutex.Lock()
	defer serverConnection.sendMutex.Unlock()
	bytesSent, err := Tcp.Send(serverConnection.netConn, message.Serialize(), client.config.TcpTimeoutMs)
	if err != nil {
		return Error.New("Failed to send message", err)
	}
	client.bytesSent.Add(bytesSent)
	if message.GetSyncTokenToken() != "" {
		client.syncRequestBytesSent.Add(bytesSent)
		client.syncRequestsSent.Add(1)
	} else {
		client.asyncMessageBytesSent.Add(bytesSent)
		client.asyncMessagesSent.Add(1)
	}
	return nil
}
