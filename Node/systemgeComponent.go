package Node

func (node *Node) startSystemgeComponent() error {
	for topic := range node.GetSystemgeComponent().GetAsyncMessageHandlers() {
		node.subscribeLoop(topic)
	}
	for topic := range node.GetSystemgeComponent().GetSyncMessageHandlers() {
		node.subscribeLoop(topic)
	}
	node.systemgeStarted = true
	return nil
}

func (node *Node) stopSystemgeComponent() error {
	node.removeAllBrokerConnections()
	node.systemgeStarted = false
	return nil
}
