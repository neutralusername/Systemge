package Node

import (
	"net"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

// outgoing connection to other nodes
// they are used to send async and sync requests and receive sync responses for their corresponding requests
type outgoingConnection struct {
	netConn        net.Conn
	endpointConfig *Config.TcpEndpoint
	name           string
	receiveMutex   sync.Mutex
	sendMutex      sync.Mutex
	topics         []string
}

func newOutgoingConnection(netConn net.Conn, endpoint *Config.TcpEndpoint, name string, topics []string) *outgoingConnection {
	outgoingConnection := &outgoingConnection{
		netConn:        netConn,
		endpointConfig: endpoint,
		name:           name,
		topics:         topics,
	}
	return outgoingConnection
}

func (systemge *systemgeComponent) removeOutgoingConnection(outgoingConnection *outgoingConnection) {
	systemge.outgoingConnectionMutex.Lock()
	defer systemge.outgoingConnectionMutex.Unlock()
	delete(systemge.outgoingConnections, outgoingConnection.name)
	for _, topic := range outgoingConnection.topics {
		topicResolutions := systemge.topicResolutions[topic]
		if topicResolutions != nil {
			delete(topicResolutions, outgoingConnection.name)
			if len(topicResolutions) == 0 {
				delete(systemge.topicResolutions, topic)
			}
		}
	}
}

func (systemge *systemgeComponent) addOutgoingConnection(outgoingConn *outgoingConnection) error {
	systemge.outgoingConnectionMutex.Lock()
	defer systemge.outgoingConnectionMutex.Unlock()
	if systemge.outgoingConnections[outgoingConn.name] != nil {
		outgoingConn.netConn.Close()
		return Error.New("Node connection already exists", nil)
	}
	systemge.outgoingConnections[outgoingConn.name] = outgoingConn
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
func (systemge *systemgeComponent) receiveFromOutgoingConnection(nodeConnection *outgoingConnection) ([]byte, error) {
	nodeConnection.receiveMutex.Lock()
	defer nodeConnection.receiveMutex.Unlock()
	messageBytes, byteCountReceived, err := Tcp.Receive(nodeConnection.netConn, 0, systemge.config.IncomingMessageByteLimit)
	if err != nil {
		return nil, Error.New("Failed to receive message", err)
	}
	systemge.bytesReceived.Add(byteCountReceived)
	systemge.incomingSyncResponseBytesReceived.Add(byteCountReceived)
	return messageBytes, nil
}
