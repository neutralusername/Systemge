package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tcp"
)

func (node *Node) AddSyncTopicRemotely(brokerConfigEndpoint Tcp.Endpoint, topics ...string) error {
	payload := ""
	for _, topic := range topics {
		payload += topic + "|"
	}
	if len(payload) == 0 {
		return Error.New("no topics provided", nil)
	}
	netConn, err := brokerConfigEndpoint.Dial()
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	_, err = Tcp.Exchange(netConn, Message.NewAsync("addSyncTopics", node.GetName(), payload), node.GetSystemgeComponent().GetSystemgeComponentConfig().TcpTimeoutMs)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	return nil
}

func (node *Node) RemoveSyncTopicRemotely(brokerConfigEndpoint Tcp.Endpoint, topics ...string) error {
	payload := ""
	for _, topic := range topics {
		payload += topic + "|"
	}
	if len(payload) == 0 {
		return Error.New("no topics provided", nil)
	}
	netConn, err := brokerConfigEndpoint.Dial()
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	_, err = Tcp.Exchange(netConn, Message.NewAsync("removeSyncTopics", node.GetName(), payload), node.GetSystemgeComponent().GetSystemgeComponentConfig().TcpTimeoutMs)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	return nil
}

func (node *Node) AddAsyncTopicRemotely(brokerConfigEndpoint Tcp.Endpoint, topics ...string) error {
	payload := ""
	for _, topic := range topics {
		payload += topic + "|"
	}
	if len(payload) == 0 {
		return Error.New("no topics provided", nil)
	}
	netConn, err := brokerConfigEndpoint.Dial()
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	_, err = Tcp.Exchange(netConn, Message.NewAsync("addAsyncTopics", node.GetName(), payload), node.GetSystemgeComponent().GetSystemgeComponentConfig().TcpTimeoutMs)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	return nil
}

func (node *Node) RemoveAsyncTopicRemotely(brokerConfigEndpoint Tcp.Endpoint, topics ...string) error {
	payload := ""
	for _, topic := range topics {
		payload += topic + "|"
	}
	if len(payload) == 0 {
		return Error.New("no topics provided", nil)
	}
	netConn, err := brokerConfigEndpoint.Dial()
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	_, err = Tcp.Exchange(netConn, Message.NewAsync("removeAsyncTopics", node.GetName(), payload), node.GetSystemgeComponent().GetSystemgeComponentConfig().TcpTimeoutMs)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	return nil
}
