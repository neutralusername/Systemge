package Node

func (node *Node) RetrieveWebsocketBytesSentCounter() uint64 {
	if websocket := node.websocket; websocket != nil {
		return websocket.bytesSentCounter.Swap(0)
	}
	return 0
}
func (node *Node) GetWebsocketBytesSentCounter() uint64 {
	if websocket := node.websocket; websocket != nil {
		return websocket.bytesSentCounter.Load()
	}
	return 0
}

func (node *Node) RetrieveWebsocketBytesReceivedCounter() uint64 {
	if websocket := node.websocket; websocket != nil {
		return websocket.bytesReceivedCounter.Swap(0)
	}
	return 0
}
func (node *Node) GetWebsocketBytesReceivedCounter() uint64 {
	if websocket := node.websocket; websocket != nil {
		return websocket.bytesReceivedCounter.Load()
	}
	return 0
}

func (node *Node) RetrieveWebsocketIncomingMessageCounter() uint32 {
	if websocket := node.websocket; websocket != nil {
		return websocket.incomingMessageCounter.Swap(0)
	}
	return 0
}
func (node *Node) GetWebsocketIncomingMessageCounter() uint32 {
	if websocket := node.websocket; websocket != nil {
		return websocket.incomingMessageCounter.Load()
	}
	return 0
}

func (node *Node) RetrieveWebsocketOutgoingMessageCounter() uint32 {
	if websocket := node.websocket; websocket != nil {
		return websocket.outgoigMessageCounter.Swap(0)
	}
	return 0
}
func (node *Node) GetWebsocketOutgoingMessageCounter() uint32 {
	if websocket := node.websocket; websocket != nil {
		return websocket.outgoigMessageCounter.Load()
	}
	return 0
}
