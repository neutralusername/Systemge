package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/TcpEndpoint"
	"Systemge/Utilities"
)

func (node *Node) AddSyncTopicRemotely(brokerConfigEndpoint TcpEndpoint.TcpEndpoint, topics ...string) error {
	payload := ""
	for _, topic := range topics {
		payload += topic + "|"
	}
	if len(payload) == 0 {
		return Error.New("no topics provided", nil)
	}
	netConn, err := brokerConfigEndpoint.TlsDial()
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	_, err = Utilities.TcpExchange(netConn, Message.NewAsync("addSyncTopics", node.GetName(), payload), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	return nil
}

func (node *Node) RemoveSyncTopicRemotely(brokerConfigEndpoint TcpEndpoint.TcpEndpoint, topics ...string) error {
	payload := ""
	for _, topic := range topics {
		payload += topic + "|"
	}
	if len(payload) == 0 {
		return Error.New("no topics provided", nil)
	}
	netConn, err := brokerConfigEndpoint.TlsDial()
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	_, err = Utilities.TcpExchange(netConn, Message.NewAsync("removeSyncTopics", node.GetName(), payload), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	return nil
}

func (node *Node) AddAsyncTopicRemotely(brokerConfigEndpoint TcpEndpoint.TcpEndpoint, topics ...string) error {
	payload := ""
	for _, topic := range topics {
		payload += topic + "|"
	}
	if len(payload) == 0 {
		return Error.New("no topics provided", nil)
	}
	netConn, err := brokerConfigEndpoint.TlsDial()
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	_, err = Utilities.TcpExchange(netConn, Message.NewAsync("addAsyncTopics", node.GetName(), payload), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	return nil
}

func (node *Node) RemoveAsyncTopicRemotely(brokerConfigEndpoint TcpEndpoint.TcpEndpoint, topics ...string) error {
	payload := ""
	for _, topic := range topics {
		payload += topic + "|"
	}
	if len(payload) == 0 {
		return Error.New("no topics provided", nil)
	}
	netConn, err := brokerConfigEndpoint.TlsDial()
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	_, err = Utilities.TcpExchange(netConn, Message.NewAsync("removeAsyncTopics", node.GetName(), payload), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("failed exchanging messages with broker", err)
	}
	return nil
}
