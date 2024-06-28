package Resolver

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
	"strings"
)

func (resolver *Resolver) handleResolverConnections() {
	for resolver.IsStarted() {
		netConn, err := resolver.tlsResolverListener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				resolver.logger.Log(Error.New("Failed to accept connection request", err).Error())
			}
			continue
		}
		go resolver.handleResolverConnection(netConn)
	}
}

func (resolver *Resolver) handleResolverConnection(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		resolver.logger.Log(Error.New("", err).Error())
		return
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "resolve" || message.GetOrigin() == "" {
		resolver.logger.Log(Error.New("Invalid connection request", nil).Error())
		return
	}
	resolver.mutex.Lock()
	broker, ok := resolver.registeredTopics[message.GetPayload()]
	resolver.mutex.Unlock()
	if !ok {
		resolver.logger.Log(Error.New("Topic not found: \""+message.GetPayload()+"\" by "+message.GetOrigin(), nil).Error())
		return
	}
	err = Utilities.TcpSend(netConn, Message.NewAsync("resolution", resolver.GetName(), broker.resolution.Marshal()).Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		resolver.logger.Log(Error.New("Failed to send resolution to \""+message.GetOrigin()+"\"", err).Error())
		return
	}
}
