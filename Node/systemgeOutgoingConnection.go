package Node

import (
	"net"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

// outgoing connection to other nodes
// they are used to send async and sync requests and receive sync responses for their corresponding requests
type outgoingConnection struct {
	netConn          net.Conn
	endpointConfig   *Config.TcpEndpoint
	name             string
	sendMutex        sync.Mutex
	topics           []string
	transient        bool
	rateLimiterBytes *Tools.RateLimiter
	rateLimiterMsgs  *Tools.RateLimiter
	tcpBuffer        []byte
}

func (outgoingConnection *outgoingConnection) receiveMessage(bufferSize uint32, incomingMessageByteLimit uint64) ([]byte, error) {
	completedMsgBytes := []byte{}
	for {
		if incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > incomingMessageByteLimit {
			return nil, Error.New("Incoming message byte limit exceeded", nil)
		}
		for i, b := range outgoingConnection.tcpBuffer {
			if b == Tcp.ENDOFMESSAGE {
				completedMsgBytes = append(completedMsgBytes, outgoingConnection.tcpBuffer[:i]...)
				outgoingConnection.tcpBuffer = outgoingConnection.tcpBuffer[i+1:]
				if incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > incomingMessageByteLimit {
					return nil, Error.New("Incoming message byte limit exceeded", nil)
				}
				return completedMsgBytes, nil
			}
		}
		completedMsgBytes = append(completedMsgBytes, outgoingConnection.tcpBuffer...)
		outgoingConnection.tcpBuffer = nil
		receivedMessageBytes, _, err := Tcp.Receive(outgoingConnection.netConn, 0, bufferSize)
		if err != nil {
			return nil, Error.New("Failed to refill tcp buffer", err)
		}
		outgoingConnection.tcpBuffer = append(outgoingConnection.tcpBuffer, receivedMessageBytes...)
	}
}

func (systemge *systemgeComponent) newOutgoingConnection(netConn net.Conn, endpoint *Config.TcpEndpoint, name string, topics []string) *outgoingConnection {
	outgoingConnection := &outgoingConnection{
		netConn:        netConn,
		endpointConfig: endpoint,
		name:           name,
		topics:         topics,
	}
	if systemge.config.OutgoingConnectionRateLimiterBytes != nil {
		outgoingConnection.rateLimiterBytes = Tools.NewRateLimiter(systemge.config.OutgoingConnectionRateLimiterBytes)
	}
	if systemge.config.OutgoingConnectionRateLimiterMsgs != nil {
		outgoingConnection.rateLimiterMsgs = Tools.NewRateLimiter(systemge.config.OutgoingConnectionRateLimiterMsgs)
	}
	return outgoingConnection
}

func (systemge *systemgeComponent) addOutgoingConnection(outgoingConn *outgoingConnection) error {
	systemge.outgoingConnectionMutex.Lock()
	defer systemge.outgoingConnectionMutex.Unlock()
	if systemge.outgoingConnections[outgoingConn.endpointConfig.Address] != nil {
		return Error.New("Node connection already exists", nil)
	}
	if !*systemge.currentlyInOutgoingConnectionLoop[outgoingConn.endpointConfig.Address] {
		return Error.New("Connection loop cancelled", nil)
	}
	systemge.outgoingConnections[outgoingConn.endpointConfig.Address] = outgoingConn
	for _, topic := range outgoingConn.topics {
		topicResolutions := systemge.topicResolutions[topic]
		if topicResolutions == nil {
			topicResolutions = make(map[string]*outgoingConnection)
			systemge.topicResolutions[topic] = topicResolutions
		}
		topicResolutions[outgoingConn.name] = outgoingConn
	}
	return nil
}

// async messages and sync requests are sent to outgoing connections
func (systemge *systemgeComponent) messageOutgoingConnection(outgoingConnection *outgoingConnection, message *Message.Message) error {
	outgoingConnection.sendMutex.Lock()
	defer outgoingConnection.sendMutex.Unlock()
	bytesSent, err := Tcp.Send(outgoingConnection.netConn, message.Serialize(), systemge.config.TcpTimeoutMs)
	if err != nil {
		return Error.New("Failed to send message", err)
	}
	systemge.bytesSent.Add(bytesSent)
	if message.GetSyncTokenToken() != "" {
		systemge.outgoingSyncRequestBytesSent.Add(bytesSent)
		systemge.outgoingSyncRequests.Add(1)
	} else {
		systemge.outgoingAsyncMessageBytesSent.Add(bytesSent)
		systemge.outgoingAsyncMessages.Add(1)
	}
	return nil
}
