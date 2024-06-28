package Resolver

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Resolution"
	"Systemge/Utilities"
	"net"
	"strings"
)

func (resolver *Resolver) handleConfigConnections() {
	for resolver.IsStarted() {
		netConn, err := resolver.tlsConfigListener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				resolver.logger.Log(Error.New("Failed to accept connection request", err).Error())
			}
			continue
		}
		go resolver.handleConfigConnection(netConn)
	}
}

func (resolver *Resolver) handleConfigConnection(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		resolver.logger.Log(Error.New("failed to receive message", err).Error())
		return
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetOrigin() == "" {
		resolver.logger.Log(Error.New("Invalid connection request \""+string(messageBytes)+"\"", nil).Error())
		return
	}
	switch message.GetTopic() {
	case "addTopic":
		segments := strings.Split(message.GetPayload(), " ")
		if len(segments) != 2 {
			err = Error.New("Invalid payload", nil)
			break
		}
		err = resolver.AddTopic(*Resolution.Unmarshal(segments[0]), segments[1])
	case "removeTopic":
		segments := strings.Split(message.GetPayload(), " ")
		if len(segments) != 1 {
			err = Error.New("Invalid payload", nil)
			break
		}
		resolver.RemoveTopic(segments[0])
	default:
		err = Error.New("Invalid config request", nil)
	}
	if err != nil {
		resolver.logger.Log(Error.New("Failed to handle config request \""+message.GetTopic()+"\"", err).Error())
		return
	}
	err = Utilities.TcpSend(netConn, Message.NewAsync("success", resolver.GetName(), "").Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		resolver.logger.Log(Error.New("Failed to send success message", err).Error())
		return
	}
}
