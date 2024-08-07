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
type incomingConnection struct {
	netConn          net.Conn
	name             string
	sendMutex        sync.Mutex
	rateLimiterBytes *Tools.RateLimiter
	rateLimiterMsgs  *Tools.RateLimiter
	tcpBuffer        []byte

	stopChannel chan bool //closing of this channel indicates that the incoming connection has finished its ongoing tasks.
}

func (systemge *systemgeComponent) newIncomingConnection(netConn net.Conn, name string) *incomingConnection {
	nodeConnection := &incomingConnection{
		netConn:     netConn,
		name:        name,
		stopChannel: make(chan bool),
	}
	if systemge.config.IncomingConnectionRateLimiterBytes != nil {
		nodeConnection.rateLimiterBytes = Tools.NewRateLimiter(systemge.config.IncomingConnectionRateLimiterBytes)
	}
	if systemge.config.IncomingConnectionRateLimiterMsgs != nil {
		nodeConnection.rateLimiterMsgs = Tools.NewRateLimiter(systemge.config.IncomingConnectionRateLimiterMsgs)
	}
	return nodeConnection
}

func (incomingConnection *incomingConnection) receiveMessage(bufferSize uint32, incomingMessageByteLimit uint64) ([]byte, error) {
	completedMsgBytes := []byte{}
	for {
		if incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > incomingMessageByteLimit {
			return nil, Error.New("Incoming message byte limit exceeded", nil)
		}
		for i, b := range incomingConnection.tcpBuffer {
			if b == Tcp.ENDOFMESSAGE {
				completedMsgBytes = append(completedMsgBytes, incomingConnection.tcpBuffer[:i]...)
				incomingConnection.tcpBuffer = incomingConnection.tcpBuffer[i+1:]
				if incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > incomingMessageByteLimit {
					return nil, Error.New("Incoming message byte limit exceeded", nil)
				}
				return completedMsgBytes, nil
			}
		}
		completedMsgBytes = append(completedMsgBytes, incomingConnection.tcpBuffer...)
		incomingConnection.tcpBuffer = nil
		receivedMessageBytes, _, err := Tcp.Receive(incomingConnection.netConn, 0, bufferSize)
		if err != nil {
			return nil, Error.New("Failed to refill tcp buffer", err)
		}
		incomingConnection.tcpBuffer = append(incomingConnection.tcpBuffer, receivedMessageBytes...)
	}
}

// sync responses are sent to incoming connections
func (systemge *systemgeComponent) messageIncomingConnection(incomingConnection *incomingConnection, message *Message.Message) error {
	incomingConnection.sendMutex.Lock()
	defer incomingConnection.sendMutex.Unlock()
	if message.GetSyncTokenToken() == "" {
		return Error.New("Cannot send async message to incoming node connection", nil)
	}
	bytesSent, err := Tcp.Send(incomingConnection.netConn, message.Serialize(), systemge.config.TcpTimeoutMs)
	if err != nil {
		return Error.New("Failed to send message", err)
	}
	systemge.bytesSent.Add(bytesSent)
	systemge.outgoingSyncResponseBytesSent.Add(bytesSent)
	systemge.outgoingSyncResponses.Add(1)
	return nil
}
