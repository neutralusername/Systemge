package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
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
	err := Utilities.TcpSend(nodeConnection.netConn, message.Serialize(), broker.config.TcpTimeoutMs)
	if err != nil {
		return Error.New("Failed to send message", err)
	}
	return nil
}

func (nodeConnection *nodeConnection) receive() (*Message.Message, error) {
	nodeConnection.receiveMutex.Lock()
	defer nodeConnection.receiveMutex.Unlock()
	messageBytes, err := Utilities.TcpReceive(nodeConnection.netConn, 0)
	if err != nil {
		return nil, Error.New("Failed to receive message", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil {
		return nil, Error.New("Failed to deserialize message \""+string(messageBytes)+"\"", nil)
	}
	return message, nil
}

func (broker *Broker) removeNodeConnection(nodeConnection *nodeConnection) {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
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
		go broker.removeNodeConnection(nodeConnection)
	}
}
