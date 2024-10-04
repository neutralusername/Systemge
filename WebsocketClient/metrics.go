package WebsocketClient

import "github.com/neutralusername/Systemge/Metrics"

func (client *WebsocketClient) GetBytesSent() uint64 {
	return client.bytesSent.Swap(0)
}
func (client *WebsocketClient) CheckBytesSent() uint64 {
	return client.bytesSent.Load()
}

func (client *WebsocketClient) GetBytesReceived() uint64 {
	return client.bytesReceived.Swap(0)
}
func (client *WebsocketClient) CheckBytesReceived() uint64 {
	return client.bytesReceived.Load()
}

func (client *WebsocketClient) GetMessagesSent() uint64 {
	return client.messagesSent.Swap(0)
}
func (client *WebsocketClient) CheckMessagesSent() uint64 {
	return client.messagesSent.Load()
}

func (client *WebsocketClient) GetMessagesReceived() uint64 {
	return client.messagesReceived.Swap(0)
}
func (client *WebsocketClient) CheckMessagesReceived() uint64 {
	return client.messagesReceived.Load()
}

func (client *WebsocketClient) GetInvalidMessagesReceived() uint64 {
	return client.invalidMessagesReceived.Swap(0)
}
func (client *WebsocketClient) CheckInvalidMessagesReceived() uint64 {
	return client.invalidMessagesReceived.Load()
}

func (client *WebsocketClient) GetRejectedMessagesReceived() uint64 {
	return client.rejectedMessagesReceived.Swap(0)
}
func (client *WebsocketClient) CheckRejectedMessagesReceived() uint64 {
	return client.rejectedMessagesReceived.Load()
}

func (client *WebsocketClient) GetMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketClient": Metrics.New(map[string]uint64{
			"bytesSent":                client.GetBytesSent(),
			"bytesReceived":            client.GetBytesReceived(),
			"messagesSent":             client.GetMessagesSent(),
			"messagesReceived":         client.GetMessagesReceived(),
			"invalidMessagesReceived":  client.GetInvalidMessagesReceived(),
			"rejectedMessagesReceived": client.GetRejectedMessagesReceived(),
		}),
	}
}

func (client *WebsocketClient) CheckMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketClient": Metrics.New(map[string]uint64{
			"bytesSent":                client.CheckBytesSent(),
			"bytesReceived":            client.CheckBytesReceived(),
			"messagesSent":             client.CheckMessagesSent(),
			"messagesReceived":         client.CheckMessagesReceived(),
			"invalidMessagesReceived":  client.CheckInvalidMessagesReceived(),
			"rejectedMessagesReceived": client.CheckRejectedMessagesReceived(),
		}),
	}
}
