package TcpSystemgeListener

import (
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/TcpSystemgeConnection"
)

func (listener *TcpSystemgeListener) AcceptConnection(serverName string, connectionConfig *Config.TcpSystemgeConnection) (SystemgeConnection.SystemgeConnection, error) {
	netConn, err := listener.listener.Accept()
	listener.connectionId++
	connectionId := listener.connectionId
	listener.connectionAttempts.Add(1)
	if err != nil {
		listener.failedConnectionAttempts.Add(1)
		return nil, Error.New("Failed to accept connection #"+Helpers.Uint32ToString(connectionId), err)
	}
	ip, _, _ := net.SplitHostPort(netConn.RemoteAddr().String())
	if listener.ipRateLimiter != nil && !listener.ipRateLimiter.RegisterConnectionAttempt(ip) {
		listener.rejectedConnectionAttempts.Add(1)
		netConn.Close()
		return nil, Error.New("Rejected connection #"+Helpers.Uint32ToString(connectionId)+" due to rate limiting", nil)
	}
	if listener.blacklist != nil && listener.blacklist.Contains(ip) {
		listener.rejectedConnectionAttempts.Add(1)
		netConn.Close()
		return nil, Error.New("Rejected connection #"+Helpers.Uint32ToString(connectionId)+" due to blacklist", nil)
	}
	if listener.whitelist != nil && listener.whitelist.ElementCount() > 0 && !listener.whitelist.Contains(ip) {
		listener.rejectedConnectionAttempts.Add(1)
		netConn.Close()
		return nil, Error.New("Rejected connection #"+Helpers.Uint32ToString(connectionId)+" due to whitelist", nil)
	}
	connection, err := listener.serverHandshake(connectionConfig, serverName, netConn)
	if err != nil {
		listener.rejectedConnectionAttempts.Add(1)
		netConn.Close()
		return nil, Error.New("Rejected connection #"+Helpers.Uint32ToString(connectionId)+" due to handshake failure", err)
	}
	listener.acceptedConnectionAttempts.Add(1)
	return connection, nil
}

func (listener *TcpSystemgeListener) serverHandshake(connectionConfig *Config.TcpSystemgeConnection, serverName string, netConn net.Conn) (*TcpSystemgeConnection.TcpSystemgeConnection, error) {
	messageReceiver := TcpSystemgeConnection.NewBufferedMessageReceiver(netConn, connectionConfig.IncomingMessageByteLimit, connectionConfig.TcpReceiveTimeoutMs, connectionConfig.TcpBufferBytes)
	messageBytes, err := messageReceiver.ReceiveNextMessage()
	if err != nil {
		return nil, Error.New("Failed to receive \""+Message.TOPIC_NAME+"\" message", err)
	}
	if len(messageBytes) == 0 {
		return nil, Error.New("Received empty message", nil)
	}
	filteresMessageBytes := []byte{}
	for _, b := range messageBytes {
		if b == Tcp.HEARTBEAT {
			continue
		}
		if b == Tcp.ENDOFMESSAGE {
			continue
		}
		filteresMessageBytes = append(filteresMessageBytes, b)
	}
	message, err := Message.Deserialize(filteresMessageBytes, "")
	if err != nil {
		return nil, Error.New("Failed to deserialize \""+Message.TOPIC_NAME+"\" message", err)
	}
	if message.GetTopic() != Message.TOPIC_NAME {
		return nil, Error.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+Message.TOPIC_NAME+"\"", nil)
	}
	if int(listener.config.MaxClientNameLength) > 0 && len(message.GetPayload()) > int(listener.config.MaxClientNameLength) {
		return nil, Error.New("Received client name \""+message.GetPayload()+"\" exceeds maximum size of "+Helpers.Uint64ToString(listener.config.MaxClientNameLength), nil)
	}
	if message.GetPayload() == "" {
		return nil, Error.New("Received empty payload in \""+Message.TOPIC_NAME+"\" message", nil)
	}
	_, err = Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_NAME, serverName).Serialize(), connectionConfig.TcpSendTimeoutMs)
	if err != nil {
		return nil, Error.New("Failed to send \""+Message.TOPIC_NAME+"\" message", err)
	}
	return TcpSystemgeConnection.New(message.GetPayload(), connectionConfig, netConn, messageReceiver), nil
}
