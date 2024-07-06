package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/TcpEndpoint"
	"Systemge/Utilities"
)

func (node *Node) AddSyncTopicRemotely(brokerConfigEndpoint TcpEndpoint.TcpEndpoint, topic string) error {
	netConn, err := brokerConfigEndpoint.TlsDial()
	if err != nil {
		return Error.New("Error dialing broker", err)
	}
	defer netConn.Close()
	_, err = Utilities.TcpExchange(netConn, Message.NewAsync("addSyncTopic", node.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}

func (node *Node) RemoveSyncTopicRemotely(brokerConfigEndpoint TcpEndpoint.TcpEndpoint, topic string) error {
	netConn, err := brokerConfigEndpoint.TlsDial()
	if err != nil {
		return Error.New("Error dialing broker", err)
	}
	defer netConn.Close()
	_, err = Utilities.TcpExchange(netConn, Message.NewAsync("removeSyncTopic", node.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}

func (node *Node) AddAsyncTopicRemotely(brokerConfigEndpoint TcpEndpoint.TcpEndpoint, topic string) error {
	netConn, err := brokerConfigEndpoint.TlsDial()
	if err != nil {
		return Error.New("Error dialing broker", err)
	}
	defer netConn.Close()
	_, err = Utilities.TcpExchange(netConn, Message.NewAsync("addAsyncTopic", node.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}

func (node *Node) RemoveAsyncTopicRemotely(brokerConfigEndpoint TcpEndpoint.TcpEndpoint, topic string) error {
	netConn, err := brokerConfigEndpoint.TlsDial()
	if err != nil {
		return Error.New("Error dialing broker", err)
	}
	defer netConn.Close()
	_, err = Utilities.TcpExchange(netConn, Message.NewAsync("removeAsyncTopic", node.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}
