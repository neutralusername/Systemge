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
	listener.acceptMutex.Lock()
	defer listener.acceptMutex.Unlock()

	tcpSystemgeConnectionId := listener.tcpSystemgeConnectionId + 1
	if event := listener.onInfo(Event.NewInfo(
		Event.AcceptingClient,
		"accepting TcpSystemgeConnection",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		listener.GetServerContext().Merge(Event.Context{
			Event.Kind:           Event.TcpSystemgeConnection,
			Event.AdditionalKind: Helpers.Uint64ToString(tcpSystemgeConnectionId),
		}),
	)); !event.IsInfo() {
		return nil, event.GetError()
	}
	listener.tcpSystemgeConnectionId++
	listener.tcpSystemgeConnectionAttemptsTotal.Add(1)

	netConn, err := listener.tcpListener.Accept()
	if err != nil {
		listener.onWarning(Event.NewWarningNoOption(
			Event.AcceptingClientFailed,
			"accepting TcpSystemgeConnection",
			listener.GetServerContext().Merge(Event.Context{
				Event.Kind:           Event.TcpSystemgeConnection,
				Event.AdditionalKind: Helpers.Uint64ToString(tcpSystemgeConnectionId),
			}),
		))
		listener.tcpSystemgeConnectionAttemptsFailed.Add(1)
		return nil, err
	}
	ip, _, _ := net.SplitHostPort(netConn.RemoteAddr().String())
	if listener.ipRateLimiter != nil && !listener.ipRateLimiter.RegisterConnectionAttempt(ip) {
		listener.tcpSystemgeConnectionAttemptsRejected.Add(1)
		netConn.Close()
		return nil, Event.New("Rejected connection #"+Helpers.Uint32ToString(tcpSystemgeConnectionId)+" due to rate limiting", nil)
	}
	if listener.blacklist != nil && listener.blacklist.Contains(ip) {
		listener.tcpSystemgeConnectionAttemptsRejected.Add(1)
		netConn.Close()
		return nil, Event.New("Rejected connection #"+Helpers.Uint32ToString(tcpSystemgeConnectionId)+" due to blacklist", nil)
	}
	if listener.whitelist != nil && listener.whitelist.ElementCount() > 0 && !listener.whitelist.Contains(ip) {
		listener.tcpSystemgeConnectionAttemptsRejected.Add(1)
		netConn.Close()
		return nil, Event.New("Rejected connection #"+Helpers.Uint32ToString(tcpSystemgeConnectionId)+" due to whitelist", nil)
	}
	connection, err := listener.serverHandshake(connectionConfig, serverName, netConn)
	if err != nil {
		listener.tcpSystemgeConnectionAttemptsRejected.Add(1)
		netConn.Close()
		return nil, Event.New("Rejected connection #"+Helpers.Uint32ToString(tcpSystemgeConnectionId)+" due to handshake failure", err)
	}
	listener.tcpSystemgeConnectionAttemptsAccepted.Add(1)
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
