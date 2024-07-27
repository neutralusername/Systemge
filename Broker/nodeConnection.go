package Broker

import (
	"net"
	"sync"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

type nodeConnection struct {
	name    string
	netConn net.Conn

	subscribedTopics map[string]bool

	sendMutex    sync.Mutex
	receiveMutex sync.Mutex
}

func (broker *Broker) newNodeConnection(name string, netConn net.Conn) *nodeConnection {
	return &nodeConnection{
		name:             name,
		netConn:          netConn,
		subscribedTopics: map[string]bool{},
	}
}

func (broker *Broker) send(nodeConnection *nodeConnection, message *Message.Message) error {
	nodeConnection.sendMutex.Lock()
	defer nodeConnection.sendMutex.Unlock()
	bytesSent, err := Tcp.Send(nodeConnection.netConn, message.Serialize(), broker.config.TcpTimeoutMs)
	if err != nil {
		return Error.New("Failed to send message", err)
	}
	broker.bytesSentCounter.Add(bytesSent)
	broker.outgoingMessageCounter.Add(1)
	return nil
}

func (broker *Broker) receive(nodeConnection *nodeConnection) (*Message.Message, error) {
	nodeConnection.receiveMutex.Lock()
	defer nodeConnection.receiveMutex.Unlock()
	messageBytes, bytesReceived, err := Tcp.Receive(nodeConnection.netConn, 0, broker.config.IncomingMessageByteLimit)
	if err != nil {
		return nil, Error.New("Failed to receive message", err)
	}
	broker.bytesReceivedCounter.Add(bytesReceived)
	broker.incomingMessageCounter.Add(1)
	message := Message.Deserialize(messageBytes)
	if message == nil {
		return nil, Error.New("Failed to deserialize message \""+string(messageBytes)+"\"", nil)
	}
	return message, nil
}

func (broker *Broker) removeNodeConnection(lock bool, nodeConnection *nodeConnection) {
	if lock {
		broker.operationMutex.Lock()
		defer broker.operationMutex.Unlock()
	}
	if nodeConnection.netConn == nil {
		return
	}
	nodeConnection.netConn.Close()
	nodeConnection.netConn = nil
	for messageType := range nodeConnection.subscribedTopics {
		delete(broker.nodeSubscriptions[messageType], nodeConnection.name)
	}
	delete(broker.nodeConnections, nodeConnection.name)
	if infoLogger := broker.node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Removed node connection \""+nodeConnection.name+"\"", nil).Error())
	}
}

func (broker *Broker) addNodeConnection(nodeConnection *nodeConnection) error {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	if broker.nodeConnections[nodeConnection.name] != nil {
		return Error.New("Node name already in use", nil)
	}
	broker.nodeConnections[nodeConnection.name] = nodeConnection
	return nil
}

func (broker *Broker) removeAllNodeConnections() {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	for _, nodeConnection := range broker.nodeConnections {
		broker.removeNodeConnection(false, nodeConnection)
	}
}
