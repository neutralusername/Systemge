package WebsocketServer

// returns the metrics without resetting the counters
func (server *WebsocketServer) CheckMetrics() map[string]map[string]uint64 {
	return map[string]map[string]uint64{
		"websocket_server": {
			"bytes_sent":         server.CheckBytesSentCounter(),
			"bytes_received":     server.CheckBytesReceivedCounter(),
			"incoming_messages":  uint64(server.CheckIncomingMessageCounter()),
			"outgoing_messages":  uint64(server.CheckOutgoingMessageCounter()),
			"active_connections": uint64(server.GetClientCount()),
		},
	}
}

// returns the metrics and resets the counters to 0
func (server *WebsocketServer) GetMetrics() map[string]map[string]uint64 {
	return map[string]map[string]uint64{
		"websocket_server": {
			"bytes_sent":         server.GetBytesSentCounter(),
			"bytes_received":     server.GetBytesReceivedCounter(),
			"incoming_messages":  uint64(server.GetIncomingMessageCounter()),
			"outgoing_messages":  uint64(server.GetOutgoingMessageCounter()),
			"active_connections": uint64(server.GetClientCount()),
		},
	}
}

func (server *WebsocketServer) GetBytesSentCounter() uint64 {
	return server.bytesSentCounter.Swap(0)
}
func (server *WebsocketServer) CheckBytesSentCounter() uint64 {
	return server.bytesSentCounter.Load()
}

func (server *WebsocketServer) GetBytesReceivedCounter() uint64 {
	return server.bytesReceivedCounter.Swap(0)
}
func (server *WebsocketServer) CheckBytesReceivedCounter() uint64 {
	return server.bytesReceivedCounter.Load()
}

func (server *WebsocketServer) GetIncomingMessageCounter() uint32 {
	return server.incomingMessageCounter.Swap(0)
}
func (server *WebsocketServer) CheckIncomingMessageCounter() uint32 {
	return server.incomingMessageCounter.Load()
}

func (server *WebsocketServer) GetOutgoingMessageCounter() uint32 {
	return server.outgoigMessageCounter.Swap(0)
}
func (server *WebsocketServer) CheckOutgoingMessageCounter() uint32 {
	return server.outgoigMessageCounter.Load()
}
