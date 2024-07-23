package Node

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tcp"
)

func (node *Node) AddSyncTopicRemotely(brokerConfigEndpoint *Config.TcpEndpoint, topics ...string) error {
	payload := ""
	for _, topic := range topics {
		payload += topic + "|"
	}
	if len(payload) == 0 {
		return Error.New("no topics provided", nil)
	}
	netConn, err := Tcp.NewEndpoint(brokerConfigEndpoint)
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	messageBytes := Message.NewAsync("addSyncTopics", node.GetName(), payload).Serialize()
	_, bytesReceived, err := Tcp.Exchange(netConn, messageBytes, node.systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, 0)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	node.systemge.bytesSentCounter.Add(uint64(len(messageBytes)))
	node.systemge.bytesReceivedCounter.Add(uint64(bytesReceived))
	return nil
}

func (node *Node) RemoveSyncTopicRemotely(brokerConfigEndpoint *Config.TcpEndpoint, topics ...string) error {
	payload := ""
	for _, topic := range topics {
		payload += topic + "|"
	}
	if len(payload) == 0 {
		return Error.New("no topics provided", nil)
	}
	netConn, err := Tcp.NewEndpoint(brokerConfigEndpoint)
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	messageBytes := Message.NewAsync("removeSyncTopics", node.GetName(), payload).Serialize()
	_, bytesReceived, err := Tcp.Exchange(netConn, messageBytes, node.systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, 0)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	node.systemge.bytesSentCounter.Add(uint64(len(messageBytes)))
	node.systemge.bytesReceivedCounter.Add(uint64(bytesReceived))
	return nil
}

func (node *Node) AddAsyncTopicRemotely(brokerConfigEndpoint *Config.TcpEndpoint, topics ...string) error {
	payload := ""
	for _, topic := range topics {
		payload += topic + "|"
	}
	if len(payload) == 0 {
		return Error.New("no topics provided", nil)
	}
	netConn, err := Tcp.NewEndpoint(brokerConfigEndpoint)
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	messageBytesCounter := Message.NewAsync("addAsyncTopics", node.GetName(), payload).Serialize()
	_, bytesReceived, err := Tcp.Exchange(netConn, messageBytesCounter, node.systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, 0)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	node.systemge.bytesSentCounter.Add(uint64(len(messageBytesCounter)))
	node.systemge.bytesReceivedCounter.Add(uint64(bytesReceived))
	return nil
}

func (node *Node) RemoveAsyncTopicRemotely(brokerConfigEndpoint *Config.TcpEndpoint, topics ...string) error {
	payload := ""
	for _, topic := range topics {
		payload += topic + "|"
	}
	if len(payload) == 0 {
		return Error.New("no topics provided", nil)
	}
	netConn, err := Tcp.NewEndpoint(brokerConfigEndpoint)
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	messageBytesCoutner := Message.NewAsync("removeAsyncTopics", node.GetName(), payload).Serialize()
	_, bytesReceived, err := Tcp.Exchange(netConn, messageBytesCoutner, node.systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, 0)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	node.systemge.bytesSentCounter.Add(uint64(len(messageBytesCoutner)))
	node.systemge.bytesReceivedCounter.Add(uint64(bytesReceived))
	return nil
}
