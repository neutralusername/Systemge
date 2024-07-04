package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
)

func (broker *Broker) handleConfigConnections() {
	for broker.IsStarted() {
		netConn, err := broker.tlsConfigListener.Accept()
		if err != nil {
			broker.tlsConfigListener.Close()
			if broker.IsStarted() {
				broker.logger.Log(Error.New("Failed to accept connection request", err).Error())
			}
			continue
		}
		go func() {
			defer netConn.Close()
			err := broker.handleConfigConnection(netConn)
			if err != nil {
				Utilities.TcpSend(netConn, Message.NewAsync("error", broker.GetName(), Error.New("failed to handle config request", err).Error()).Serialize(), DEFAULT_TCP_TIMEOUT)
			} else {
				Utilities.TcpSend(netConn, Message.NewAsync("success", broker.GetName(), "").Serialize(), DEFAULT_TCP_TIMEOUT)
			}
		}()
	}
}

func (broker *Broker) handleConfigConnection(netConn net.Conn) error {
	messageBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("failed to receive message", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetOrigin() == "" {
		return Error.New("Invalid connection request \""+string(messageBytes)+"\"", nil)
	}
	switch message.GetTopic() {
	case "addSyncTopic":
		broker.AddSyncTopics(message.GetPayload())
	case "removeSyncTopic":
		broker.RemoveSyncTopics(message.GetPayload())
	case "addAsyncTopic":
		broker.AddAsyncTopics(message.GetPayload())
	case "removeAsyncTopic":
		broker.RemoveAsyncTopics(message.GetPayload())
	default:
		return Error.New("unknown topic \""+message.GetTopic()+"\"", nil)
	}
	return nil
}
