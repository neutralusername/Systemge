package Node

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
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

	incomingAsyncMessageCounter atomic.Uint32
	incomingSyncRequestCounter  atomic.Uint32
	incomingSyncResponseCounter atomic.Uint32
	outgoingAsyncMessageCounter atomic.Uint32
	outgoingSyncRequestCounter  atomic.Uint32
	outgoingSyncResponseCounter atomic.Uint32

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
		return systemge.incomingSyncRequestCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeIncomingSyncResponseMessageCounter() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncResponseCounter.Swap(0)
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
		return systemge.outgoingSyncRequestCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeOutgoingSyncResponseMessageCounter() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncResponseCounter.Swap(0)
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
	node.systemge.mutex.Lock()
	defer node.systemge.mutex.Unlock()
	node.systemge = nil
	for _, brokerConnection := range node.systemge.brokerConnections {
		brokerConnection.closeNetConn()
		delete(node.systemge.brokerConnections, brokerConnection.endpoint.Address)
	}
	return nil
}
