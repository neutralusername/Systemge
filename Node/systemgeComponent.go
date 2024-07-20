package Node

import "Systemge/Error"

func (node *Node) startSystemgeComponent() error {
	for topic := range node.GetSystemgeComponent().GetAsyncMessageHandlers() {
		err := node.subscribeLoop(topic, 1)
		if err != nil {
			return Error.New("Failed to subscribe for topic \""+topic+"\"", err)
		}
	}
	for topic := range node.GetSystemgeComponent().GetSyncMessageHandlers() {
		err := node.subscribeLoop(topic, 1)
		if err != nil {
			return Error.New("Failed to subscribe for topic \""+topic+"\"", err)
		}
	}
	node.systemgeStarted = true
	return nil
}

func (node *Node) stopSystemgeComponent() error {
	node.removeAllBrokerConnections()
	node.systemgeStarted = false
	return nil
}
