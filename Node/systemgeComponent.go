package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"sync"
)

type systemgeComponent struct {
	application                SystemgeComponent
	mutex                      sync.Mutex
	handleSequentiallyMutex    sync.Mutex
	messagesWaitingForResponse map[string]chan *Message.Message // syncKey -> responseChannel
	brokerConnections          map[string]*brokerConnection     // brokerAddress -> brokerConnection
	topicResolutions           map[string]*brokerConnection     // topic -> brokerConnection
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
