package Node

import (
	"net"
	"sync"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

// incoming connections from other nodes
// they are used to receive async and sync requests and send sync responses for their corresponding requests
type incomingConnection struct {
	netConn      net.Conn
	name         string
	receiveMutex sync.Mutex
	sendMutex    sync.Mutex
}

func (systemge *systemgeComponent) newIncomingConnection(netConn net.Conn, name string) (*incomingConnection, error) {
	nodeConnection := &incomingConnection{
		netConn: netConn,
		name:    name,
	}
	systemge.outgoingConnectionMutex.Lock()
	defer systemge.outgoingConnectionMutex.Unlock()
	if systemge.incomingConnections[name] != nil {
		netConn.Close()
		return nil, Error.New("Node connection already exists", nil)
	}
	systemge.incomingConnections[name] = nodeConnection
	return nodeConnection, nil
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

// async messages and sync requests are received from incoming connections
func (systemge *systemgeComponent) receiveFromIncomingConnection(incomingConnection *incomingConnection) ([]byte, error) {
	incomingConnection.receiveMutex.Lock()
	defer incomingConnection.receiveMutex.Unlock()
	messageBytes, bytesReceived, err := Tcp.Receive(incomingConnection.netConn, 0, systemge.config.IncomingMessageByteLimit)
	if err != nil {
		return nil, Error.New("Failed to receive message", err)
	}
	systemge.bytesReceived.Add(bytesReceived)
	return messageBytes, nil
}
