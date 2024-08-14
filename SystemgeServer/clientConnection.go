package Node

import (
	"net"
	"sync"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

// incoming connections from other nodes
// they are used to receive async and sync requests and send sync responses for their corresponding requests
type clientConnection struct {
	netConn          net.Conn
	name             string
	sendMutex        sync.Mutex
	rateLimiterBytes *Tools.TokenBucketRateLimiter
	rateLimiterMsgs  *Tools.TokenBucketRateLimiter
	tcpBuffer        []byte

	stopChannel chan bool //closing of this channel indicates that the incoming connection has finished its ongoing tasks.
}

func (server *SystemgeServer) newClientConnection(netConn net.Conn, name string) *clientConnection {
	nodeConnection := &clientConnection{
		netConn:     netConn,
		name:        name,
		stopChannel: make(chan bool),
	}
	if server.config.RateLimterBytes != nil {
		nodeConnection.rateLimiterBytes = Tools.NewTokenBucketRateLimiter(server.config.RateLimterBytes)
	}
	if server.config.RateLimiterMessages != nil {
		nodeConnection.rateLimiterMsgs = Tools.NewTokenBucketRateLimiter(server.config.RateLimiterMessages)
	}
	return nodeConnection
}

func (clientConnection *clientConnection) receiveMessage(bufferSize uint32, incomingMessageByteLimit uint64) ([]byte, error) {
	completedMsgBytes := []byte{}
	for {
		if incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > incomingMessageByteLimit {
			return nil, Error.New("Incoming message byte limit exceeded", nil)
		}
		for i, b := range clientConnection.tcpBuffer {
			if b == Tcp.ENDOFMESSAGE {
				completedMsgBytes = append(completedMsgBytes, clientConnection.tcpBuffer[:i]...)
				clientConnection.tcpBuffer = clientConnection.tcpBuffer[i+1:]
				if incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > incomingMessageByteLimit {
					return nil, Error.New("Incoming message byte limit exceeded", nil)
				}
				return completedMsgBytes, nil
			}
		}
		completedMsgBytes = append(completedMsgBytes, clientConnection.tcpBuffer...)
		receivedMessageBytes, _, err := Tcp.Receive(clientConnection.netConn, 0, bufferSize)
		if err != nil {
			return nil, Error.New("Failed to refill tcp buffer", err)
		}
		clientConnection.tcpBuffer = receivedMessageBytes
	}
}

// sync responses are sent to incoming connections
func (server *SystemgeServer) messageIncomingConnection(clientConnection *clientConnection, message *Message.Message) error {
	clientConnection.sendMutex.Lock()
	defer clientConnection.sendMutex.Unlock()
	if message.GetSyncTokenToken() == "" {
		return Error.New("Cannot send async message to incoming node connection", nil)
	}
	bytesSent, err := Tcp.Send(clientConnection.netConn, message.Serialize(), server.config.TcpTimeoutMs)
	if err != nil {
		return Error.New("Failed to send message", err)
	}
	server.bytesSent.Add(bytesSent)
	server.syncResponseBytesSent.Add(bytesSent)
	return nil
}
