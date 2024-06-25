package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
	"strings"
)

func (server *Server) handleConfigConnections() {
	for server.IsStarted() {
		netConn, err := server.tlsConfigListener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				server.logger.Log(Error.New("Failed to accept connection request", err).Error())
			}
			continue
		}
		go func() {
			defer netConn.Close()
			err := server.handleConfigConnection(netConn)
			if err != nil {
				Utilities.TcpSend(netConn, Message.NewAsync("error", server.GetName(), Error.New("failed to handle config request", err).Error()).Serialize(), DEFAULT_TCP_TIMEOUT)
			} else {
				Utilities.TcpSend(netConn, Message.NewAsync("success", server.GetName(), "").Serialize(), DEFAULT_TCP_TIMEOUT)
			}
		}()
	}
}

func (server *Server) handleConfigConnection(netConn net.Conn) error {
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
		server.AddSyncTopics(message.GetPayload())
	case "removeSyncTopic":
		server.RemoveSyncTopics(message.GetPayload())
	case "addAsyncTopic":
		server.AddAsyncTopics(message.GetPayload())
	case "removeAsyncTopic":
		server.RemoveAsyncTopics(message.GetPayload())
	default:
		return Error.New("unknown topic \""+message.GetTopic()+"\"", nil)
	}
	return nil
}
