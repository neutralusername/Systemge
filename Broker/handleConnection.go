package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/TCP"
	"net"
	"strings"
)

func (server *Server) handleTlsConnections() {
	for server.IsStarted() {
		netConn, err := server.tlsListener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				panic(err)
			} else {
				server.Stop()
				return
			}
		}
		go func() {
			client, err := server.handleConnectionRequest(netConn)
			if err != nil {
				netConn.Close()
				server.logger.Log(Error.New("Failed to handle connection request", err).Error())
				return
			}
			server.handleClientMessages(client)
		}()
	}
}

func (server *Server) handleConnectionRequest(netConn net.Conn) (*ClientConnection, error) {
	messageBytes, err := TCP.Receive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return nil, Error.New("Failed to receive connection request", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "connect" || message.GetOrigin() == "" {
		return nil, Error.New("Invalid connection request \""+string(messageBytes)+"\"", nil)
	}
	client := NewClienn(message.GetOrigin(), netConn)
	err = server.addClient(client)
	if err != nil {
		return nil, err
	}
	err = client.send(Message.NewAsync("connected", server.name, ""))
	if err != nil {
		return nil, Error.New("Failed to send connection response to client \""+client.name+"\"", err)
	}
	return client, nil
}
