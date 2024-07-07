package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
	"sync"
)

type nodeConnection struct {
	name         string
	netConn      net.Conn
	messageQueue chan *Message.Message

	subscribedTopics   map[string]bool
	deliverImmediately bool

	sendMutex    sync.Mutex
	receiveMutex sync.Mutex

	stopChannel chan bool
}

func (broker *Broker) newNodeConnection(name string, netConn net.Conn) *nodeConnection {
	return &nodeConnection{
		name:               name,
		netConn:            netConn,
		messageQueue:       make(chan *Message.Message, NODE_MESSAGE_QUEUE_SIZE),
		subscribedTopics:   map[string]bool{},
		deliverImmediately: broker.config.DeliverImmediately,
		stopChannel:        make(chan bool),
	}
}

func (nodeConnection *nodeConnection) send(message *Message.Message) error {
	nodeConnection.sendMutex.Lock()
	defer nodeConnection.sendMutex.Unlock()
	return Utilities.TcpSend(nodeConnection.netConn, message.Serialize(), DEFAULT_TCP_TIMEOUT)
}

func (nodeConnection *nodeConnection) receive() ([]byte, error) {
	nodeConnection.receiveMutex.Lock()
	defer nodeConnection.receiveMutex.Unlock()
	messageBytes, err := Utilities.TcpReceive(nodeConnection.netConn, 0)
	if err != nil {
		return nil, err
	}
	return messageBytes, nil
}

func (broker *Broker) disconnect(nodeConnection *nodeConnection) {
	nodeConnection.netConn.Close()
	err := broker.removeNodeConnection(nodeConnection)
	if err != nil {
		broker.logger.Log(Error.New("Error removing node \""+nodeConnection.name+"\"", err).Error())
	}
	close(nodeConnection.stopChannel)
}

func (broker *Broker) addNodeConnection(nodeConnection *nodeConnection) error {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	if broker.nodeConnections[nodeConnection.name] != nil {
		return Error.New("node with name \""+nodeConnection.name+"\" already exists", nil)
	}
	broker.nodeConnections[nodeConnection.name] = nodeConnection
	return nil
}

func (broker *Broker) removeNodeConnection(nodeConnection *nodeConnection) error {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	if broker.nodeConnections[nodeConnection.name] == nil {
		return Error.New("Subscriber with name \""+nodeConnection.name+"\" does not exist", nil)
	}
	for messageType := range nodeConnection.subscribedTopics {
		delete(broker.nodeSubscriptions[messageType], nodeConnection.name)
	}
	delete(broker.nodeConnections, nodeConnection.name)
	return nil
}

func (broker *Broker) disconnectAllNodeConnections() {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	for _, nodeConnection := range broker.nodeConnections {
		go broker.disconnect(nodeConnection)
	}
}
