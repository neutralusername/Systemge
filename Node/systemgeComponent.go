package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"sync"
	"sync/atomic"
)

type systemgeComponent struct {
	application                        SystemgeComponent
	mutex                              sync.Mutex
	handleSequentiallyMutex            sync.Mutex
	messagesWaitingForResponse         map[string]chan *Message.Message // syncKey -> responseChannel
	brokerConnections                  map[string]*brokerConnection     // brokerAddress -> brokerConnection
	topicResolutions                   map[string]*brokerConnection     // topic -> brokerConnection
	asyncMessageHandlerMutex           sync.Mutex
	syncMessageHandlerMutex            sync.Mutex
	incomingAsyncMessageCounter        atomic.Uint32
	incomingSyncRequestMessageCounter  atomic.Uint32
	incomingSyncResponseMessageCounter atomic.Uint32
	outgoingAsyncMessageCounter        atomic.Uint32
	outgoingSyncRequestMessageCounter  atomic.Uint32
	outgoingSyncResponseMessageCounter atomic.Uint32
}

func (node *Node) GetMessageCounters() (uint32, uint32, uint32, uint32, uint32, uint32) {
	if systemge := node.systemge; systemge != nil {
		incSyncRequest := node.systemge.incomingSyncRequestMessageCounter.Swap(0)
		incSyncResponse := node.systemge.incomingSyncResponseMessageCounter.Swap(0)
		incAsync := node.systemge.incomingAsyncMessageCounter.Swap(0)
		outSyncRequest := node.systemge.outgoingSyncRequestMessageCounter.Swap(0)
		outSyncResponse := node.systemge.outgoingSyncResponseMessageCounter.Swap(0)
		outAsync := node.systemge.outgoingAsyncMessageCounter.Swap(0)
		return incSyncRequest, incSyncResponse, incAsync, outSyncRequest, outSyncResponse, outAsync
	}
	return 0, 0, 0, 0, 0, 0
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
