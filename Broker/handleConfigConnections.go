package Broker

import (
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
				server.logger.Log(Utilities.NewError("Failed to accept connection request", err).Error())
			}
			continue
		}
		go func() {
			err := server.handleConfigConnection(netConn)
			netConn.Close()
			if err != nil {
				Utilities.TcpSend(netConn, Message.NewAsync("error", server.GetName(), Utilities.NewError("failed to handle config request", err).Error()).Serialize(), DEFAULT_TCP_TIMEOUT)
			} else {
				Utilities.TcpSend(netConn, Message.NewAsync("success", server.GetName(), "").Serialize(), DEFAULT_TCP_TIMEOUT)
			}
		}()
	}
}

func (server *Server) handleConfigConnection(netConn net.Conn) error {
	messageBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("failed to receive message", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetOrigin() == "" {
		return Utilities.NewError("Invalid connection request \""+string(messageBytes)+"\"", nil)
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
	case "removeClient":
		server.operationMutex.Lock()
		clientConnection := server.clientConnections[message.GetPayload()]
		server.operationMutex.Unlock()
		if clientConnection == nil {
			return Utilities.NewError("Client \""+message.GetPayload()+"\" does not exist", nil)
		}
		clientConnection.disconnect()
	default:
		return Utilities.NewError("unknown topic \""+message.GetTopic()+"\"", nil)
	}
	return nil
}
