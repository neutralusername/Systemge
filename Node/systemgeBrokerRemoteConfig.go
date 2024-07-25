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
	payload = payload[:len(payload)-1]
	netConn, err := Tcp.NewEndpoint(brokerConfigEndpoint)
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	_, bytesSent, bytesReceived, err := Tcp.Exchange(netConn, Message.NewAsync("addSyncTopics", node.GetName(), payload).Serialize(), node.systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, 0)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	node.systemge.bytesSentCounter.Add(bytesSent)
	node.systemge.bytesReceivedCounter.Add(bytesReceived)
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
	payload = payload[:len(payload)-1]
	netConn, err := Tcp.NewEndpoint(brokerConfigEndpoint)
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	_, bytesSent, bytesReceived, err := Tcp.Exchange(netConn, Message.NewAsync("removeSyncTopics", node.GetName(), payload).Serialize(), node.systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, 0)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	node.systemge.bytesSentCounter.Add(bytesSent)
	node.systemge.bytesReceivedCounter.Add(bytesReceived)
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
	payload = payload[:len(payload)-1]
	netConn, err := Tcp.NewEndpoint(brokerConfigEndpoint)
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	_, bytesSent, bytesReceived, err := Tcp.Exchange(netConn, Message.NewAsync("addAsyncTopics", node.GetName(), payload).Serialize(), node.systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, 0)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	node.systemge.bytesSentCounter.Add(bytesSent)
	node.systemge.bytesReceivedCounter.Add(bytesReceived)
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
	payload = payload[:len(payload)-1]
	netConn, err := Tcp.NewEndpoint(brokerConfigEndpoint)
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	_, bytesSent, bytesReceived, err := Tcp.Exchange(netConn, Message.NewAsync("removeAsyncTopics", node.GetName(), payload).Serialize(), node.systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, 0)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	node.systemge.bytesSentCounter.Add(bytesSent)
	node.systemge.bytesReceivedCounter.Add(bytesReceived)
	return nil
}
