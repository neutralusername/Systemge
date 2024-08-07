package Node

func (node *Node) RetrieveHTTPRequestCounter() uint64 {
	if httpComponent := node.http; httpComponent != nil {
		return httpComponent.requestCounter.Swap(0)
	}
	return 0
}

func (node *Node) GetHTTPRequestCounter() uint64 {
	if httpComponent := node.http; httpComponent != nil {
		return httpComponent.requestCounter.Load()
	}
	return 0
}
