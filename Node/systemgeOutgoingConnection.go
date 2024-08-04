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
	receiveMutex     sync.Mutex
	sendMutex        sync.Mutex
	topics           []string
	transient        bool
	rateLimiterBytes *Tools.RateLimiter
	rateLimiterMsgs  *Tools.RateLimiter
}

func (node *Node) RemoveOutgoingConnection(address string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.outgoingConnectionMutex.Lock()
		defer systemge.outgoingConnectionMutex.Unlock()
		if systemge.currentlyInOutgoingConnectionLoop[address] != nil {
			*systemge.currentlyInOutgoingConnectionLoop[address] = false
			delete(systemge.currentlyInOutgoingConnectionLoop, address)
		}
		if outgoingConnection := systemge.outgoingConnections[address]; outgoingConnection != nil {
			outgoingConnection.netConn.Close()
			outgoingConnection.transient = true
		}
		return nil
	}
	return Error.New("Systemge is nil", nil)
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
		outgoingConn.netConn.Close()
		return Error.New("Node connection already exists", nil)
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

// sync responses are received by outgoing connections
func (systemge *systemgeComponent) receiveFromOutgoingConnection(outgoingConnection *outgoingConnection) ([]byte, error) {
	outgoingConnection.receiveMutex.Lock()
	defer outgoingConnection.receiveMutex.Unlock()
	messageBytes, byteCountReceived, err := Tcp.Receive(outgoingConnection.netConn, 0, systemge.config.IncomingMessageByteLimit)
	if err != nil {
		return nil, Error.New("Failed to receive message", err)
	}
	systemge.bytesReceived.Add(byteCountReceived)
	return messageBytes, nil
}
