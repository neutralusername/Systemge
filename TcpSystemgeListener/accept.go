package TcpSystemgeListener

import (
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/TcpSystemgeConnection"
)

func (listener *TcpSystemgeListener) AcceptConnection(serverName string, connectionConfig *Config.TcpSystemgeConnection) (SystemgeConnection.SystemgeConnection, error) {
	listener.connectionAttempts.Add(1)
	listener.connectionId++
	connectionId := listener.connectionId
	netConn, err := listener.listener.Accept()
	if err != nil {
		listener.onWarning(Event.NewWarningNoOption(
			Event.AcceptClientFailed,
			"accepting TcpSystemgeConnection",
			listener.GetServerContext().Merge(Event.Context{
				Event.Kind: Event.TcpSystemgeConnection,
			}),
		))
		listener.failedConnectionAttempts.Add(1)
		return nil, Event.New("Failed to accept connection #"+Helpers.Uint32ToString(connectionId), err)
	}
	ip, _, _ := net.SplitHostPort(netConn.RemoteAddr().String())
	if listener.ipRateLimiter != nil && !listener.ipRateLimiter.RegisterConnectionAttempt(ip) {
		listener.rejectedConnectionAttempts.Add(1)
		netConn.Close()
		return nil, Event.New("Rejected connection #"+Helpers.Uint32ToString(connectionId)+" due to rate limiting", nil)
	}
	if listener.blacklist != nil && listener.blacklist.Contains(ip) {
		listener.rejectedConnectionAttempts.Add(1)
		netConn.Close()
		return nil, Event.New("Rejected connection #"+Helpers.Uint32ToString(connectionId)+" due to blacklist", nil)
	}
	if listener.whitelist != nil && listener.whitelist.ElementCount() > 0 && !listener.whitelist.Contains(ip) {
		listener.rejectedConnectionAttempts.Add(1)
		netConn.Close()
		return nil, Event.New("Rejected connection #"+Helpers.Uint32ToString(connectionId)+" due to whitelist", nil)
	}
	connection, err := listener.serverHandshake(connectionConfig, serverName, netConn)
	if err != nil {
		listener.rejectedConnectionAttempts.Add(1)
		netConn.Close()
		return nil, Event.New("Rejected connection #"+Helpers.Uint32ToString(connectionId)+" due to handshake failure", err)
	}
	listener.acceptedConnectionAttempts.Add(1)
	return connection, nil
}

func (listener *TcpSystemgeListener) serverHandshake(connectionConfig *Config.TcpSystemgeConnection, serverName string, netConn net.Conn) (*TcpSystemgeConnection.TcpSystemgeConnection, error) {
	messageReceiver := TcpSystemgeConnection.NewBufferedMessageReceiver(netConn, connectionConfig.IncomingMessageByteLimit, connectionConfig.TcpReceiveTimeoutMs, connectionConfig.TcpBufferBytes)
	messageBytes, err := messageReceiver.ReceiveNextMessage()
	if err != nil {
		return nil, Event.New("Failed to receive \""+Message.TOPIC_NAME+"\" message", err)
	}
	if len(messageBytes) == 0 {
		return nil, Event.New("Received empty message", nil)
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
		return nil, Event.New("Failed to deserialize \""+Message.TOPIC_NAME+"\" message", err)
	}
	if message.GetTopic() != Message.TOPIC_NAME {
		return nil, Event.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+Message.TOPIC_NAME+"\"", nil)
	}
	if int(listener.config.MaxClientNameLength) > 0 && len(message.GetPayload()) > int(listener.config.MaxClientNameLength) {
		return nil, Event.New("Received client name \""+message.GetPayload()+"\" exceeds maximum size of "+Helpers.Uint64ToString(listener.config.MaxClientNameLength), nil)
	}
	if message.GetPayload() == "" {
		return nil, Event.New("Received empty payload in \""+Message.TOPIC_NAME+"\" message", nil)
	}
	_, err = Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_NAME, serverName).Serialize(), connectionConfig.TcpSendTimeoutMs)
	if err != nil {
		return nil, Event.New("Failed to send \""+Message.TOPIC_NAME+"\" message", err)
	}
	return TcpSystemgeConnection.New(message.GetPayload(), connectionConfig, netConn, messageReceiver), nil
}
