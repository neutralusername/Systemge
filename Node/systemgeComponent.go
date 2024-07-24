package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"sync"
	"sync/atomic"
)

type systemgeComponent struct {
	application                SystemgeComponent
	mutex                      sync.Mutex
	handleSequentiallyMutex    sync.Mutex
	messagesWaitingForResponse map[string]chan *Message.Message // syncKey -> responseChannel
	brokerConnections          map[string]*brokerConnection     // brokerAddress -> brokerConnection
	topicResolutions           map[string]*brokerConnection     // topic -> brokerConnection
	asyncMessageHandlerMutex   sync.Mutex
	syncMessageHandlerMutex    sync.Mutex

	incomingAsyncMessageCounter        atomic.Uint32
	incomingSyncRequestMessageCounter  atomic.Uint32
	incomingSyncResponseMessageCounter atomic.Uint32
	outgoingAsyncMessageCounter        atomic.Uint32
	outgoingSyncRequestMessageCounter  atomic.Uint32
	outgoingSyncResponseMessageCounter atomic.Uint32

	bytesReceivedCounter atomic.Uint64
	bytesSentCounter     atomic.Uint64
}

func (node *Node) RetrieveSystemgeBytesReceivedCounter() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.bytesReceivedCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeBytesSentCounter() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.bytesSentCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeIncomingAsyncMessageCounter() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingAsyncMessageCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeIncomingSyncRequestMessageCounter() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncRequestMessageCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeIncomingSyncResponseMessageCounter() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncResponseMessageCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeOutgoingAsyncMessageCounter() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingAsyncMessageCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeOutgoingSyncRequestMessageCounter() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncRequestMessageCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeOutgoingSyncResponseMessageCounter() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncResponseMessageCounter.Swap(0)
	}
	return 0
}

func (node *Node) startSystemgeComponent() error {
	node.systemge = &systemgeComponent{
		application:                node.application.(SystemgeComponent),
		messagesWaitingForResponse: make(map[string]chan *Message.Message),
		brokerConnections:          make(map[string]*brokerConnection),
		topicResolutions:           make(map[string]*brokerConnection),
	}
	for topic := range node.systemge.application.GetAsyncMessageHandlers() {
		err := node.subscribeLoop(topic, 1)
		if err != nil {
			return Error.New("Failed to subscribe for topic \""+topic+"\"", err)
		}
	}
	for topic := range node.systemge.application.GetSyncMessageHandlers() {
		err := node.subscribeLoop(topic, 1)
		if err != nil {
			return Error.New("Failed to subscribe for topic \""+topic+"\"", err)
		}
	}
	return nil
}

func (node *Node) stopSystemgeComponent() error {
	systemge := node.systemge
	node.systemge = nil
	systemge.removeAllBrokerConnections()
	return nil
}
