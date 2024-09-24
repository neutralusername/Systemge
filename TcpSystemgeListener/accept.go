package TcpSystemgeListener

import (
	"errors"
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/TcpSystemgeConnection"
)

func (listener *TcpSystemgeListener) AcceptConnection(connectionConfig *Config.TcpSystemgeConnection) (SystemgeConnection.SystemgeConnection, error) {
	listener.acceptMutex.Lock()
	defer listener.acceptMutex.Unlock()

	if event := listener.onEvent(Event.NewInfo(
		Event.AcceptingClient,
		"accepting TcpSystemgeConnection",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		listener.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.TcpSystemgeListenerAcceptRoutine,
			Event.ClientType:   Event.TcpSystemgeConnection,
		}),
	)); !event.IsInfo() {
		return nil, event.GetError()
	}
	listener.tcpSystemgeConnectionAttemptsTotal.Add(1)

	netConn, err := listener.tcpListener.Accept()
	if err != nil {
		listener.onEvent(Event.NewWarningNoOption(
			Event.AcceptingClientFailed,
			"accepting TcpSystemgeConnection",
			listener.GetServerContext().Merge(Event.Context{
				Event.Circumstance: Event.TcpSystemgeListenerAcceptRoutine,
				Event.ClientType:   Event.TcpSystemgeConnection,
			}),
		))
		listener.tcpSystemgeConnectionAttemptsFailed.Add(1)
		return nil, err
	}

	ip, _, err := net.SplitHostPort(netConn.RemoteAddr().String())
	if err != nil {
		listener.onEvent(Event.NewWarningNoOption(
			Event.SplittingHostPortFailed,
			err.Error(),
			listener.GetServerContext().Merge(Event.Context{
				Event.Circumstance:  Event.WebsocketUpgradeRoutine,
				Event.ClientAddress: netConn.RemoteAddr().String(),
			}),
		))
		listener.tcpSystemgeConnectionAttemptsFailed.Add(1)
		return nil, err
	}

	if listener.ipRateLimiter != nil && !listener.ipRateLimiter.RegisterConnectionAttempt(ip) {
		if event := listener.onEvent(Event.NewWarning(
			Event.RateLimited,
			"tcpSystemgeConnection attempt ip rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			listener.GetServerContext().Merge(Event.Context{
				Event.Circumstance:    Event.TcpSystemgeListenerAcceptRoutine,
				Event.RateLimiterType: Event.Ip,
				Event.ClientAddress:   netConn.RemoteAddr().String(),
			}),
		)); !event.IsInfo() {
			listener.tcpSystemgeConnectionAttemptsRejected.Add(1)
			netConn.Close()
			return nil, errors.New("Rate limit exceeded")
		}
	}

	if listener.blacklist != nil && listener.blacklist.Contains(ip) {
		if event := listener.onEvent(Event.NewWarning(
			Event.Blacklisted,
			"tcpSystemgeConnection attempt ip blacklisted",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			listener.GetServerContext().Merge(Event.Context{
				Event.Circumstance:  Event.TcpSystemgeListenerAcceptRoutine,
				Event.ClientAddress: netConn.RemoteAddr().String(),
			}),
		)); !event.IsInfo() {
			listener.tcpSystemgeConnectionAttemptsRejected.Add(1)
			netConn.Close()
			return nil, errors.New("Blacklisted")
		}
	}

	if listener.whitelist != nil && listener.whitelist.ElementCount() > 0 && !listener.whitelist.Contains(ip) {
		if event := listener.onEvent(Event.NewWarning(
			Event.NotWhitelisted,
			"tcpSystemgeConnection attempt ip not whitelisted",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			listener.GetServerContext().Merge(Event.Context{
				Event.Circumstance:  Event.TcpSystemgeListenerAcceptRoutine,
				Event.ClientAddress: netConn.RemoteAddr().String(),
			}),
		)); !event.IsInfo() {
			listener.tcpSystemgeConnectionAttemptsRejected.Add(1)
			netConn.Close()
			return nil, errors.New("Not whitelisted")
		}
	}

	connection, err := listener.serverHandshake(connectionConfig, netConn)
	if err != nil {
		listener.tcpSystemgeConnectionAttemptsRejected.Add(1)
		netConn.Close()
		return nil, err
	}

	if event := listener.onEvent(Event.NewInfo(
		Event.AcceptedClient,
		"accepted TcpSystemgeConnection",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		listener.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.TcpSystemgeListenerAcceptRoutine,
			Event.ClientType:   Event.TcpSystemgeConnection,
		}),
	)); !event.IsInfo() {
		connection.Close()
		listener.tcpSystemgeConnectionAttemptsRejected.Add(1)
		return nil, event.GetError()
	}

	listener.tcpSystemgeConnectionAttemptsAccepted.Add(1)
	return connection, nil
}

func (listener *TcpSystemgeListener) serverHandshake(connectionConfig *Config.TcpSystemgeConnection, netConn net.Conn) (*TcpSystemgeConnection.TcpSystemgeConnection, error) {
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
	_, err = Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_NAME, listener.name).Serialize(), connectionConfig.TcpSendTimeoutMs)
	if err != nil {
		return nil, Event.New("Failed to send \""+Message.TOPIC_NAME+"\" message", err)
	}
	return TcpSystemgeConnection.New(message.GetPayload(), connectionConfig, netConn, messageReceiver), nil
}
