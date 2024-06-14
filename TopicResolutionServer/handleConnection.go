package TopicResolutionServer

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/TCP"
	"net"
)

func (server *Server) handleConnection(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, err := TCP.Receive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		server.logger.Log(Error.New("", err).Error())
		return
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "resolve" || message.GetOrigin() == "" {
		server.logger.Log(Error.New("Invalid connection request", nil).Error())
		return
	}
	server.mutex.Lock()
	messageBrokerAddress, ok := server.registeredTopics[message.GetPayload()]
	server.mutex.Unlock()
	if !ok {
		server.logger.Log(Error.New("Topic not found: \""+message.GetPayload()+"\" by "+message.GetOrigin(), nil).Error())
		return
	}
	err = TCP.Send(netConn, Message.NewAsync("resolution", server.name, messageBrokerAddress).Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		server.logger.Log(Error.New("Failed to send resolution to \""+message.GetOrigin()+"\"", err).Error())
	}
}
