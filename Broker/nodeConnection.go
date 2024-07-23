package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tcp"
	"net"
	"sync"
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
	err := Tcp.Send(nodeConnection.netConn, message.Serialize(), broker.config.TcpTimeoutMs)
	if err != nil {
		return Error.New("Failed to send message", err)
	}
	return nil
}

func (broker *Broker) receive(nodeConnection *nodeConnection) (*Message.Message, error) {
	nodeConnection.receiveMutex.Lock()
	defer nodeConnection.receiveMutex.Unlock()
	messageBytes, _, err := Tcp.Receive(nodeConnection.netConn, 0, broker.config.IncomingMessageByteLimit)
	if err != nil {
		return nil, Error.New("Failed to receive message", err)
	}
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
