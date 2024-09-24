package TcpSystemgeListener

import (
	"errors"
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/SystemgeConnection"
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
		Event.Context{
			Event.Circumstance: Event.TcpSystemgeListenerAcceptRoutine,
			Event.ClientType:   Event.TcpSystemgeConnection,
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}
	listener.tcpSystemgeConnectionAttemptsTotal.Add(1)

	netConn, err := listener.tcpListener.Accept()
	if err != nil {
		listener.onEvent(Event.NewWarningNoOption(
			Event.AcceptingClientFailed,
			"accepting TcpSystemgeConnection",
			Event.Context{
				Event.Circumstance: Event.TcpSystemgeListenerAcceptRoutine,
				Event.ClientType:   Event.TcpSystemgeConnection,
			},
		))
		listener.tcpSystemgeConnectionAttemptsFailed.Add(1)
		return nil, err
	}

	ip, _, err := net.SplitHostPort(netConn.RemoteAddr().String())
	if err != nil {
		listener.onEvent(Event.NewWarningNoOption(
			Event.SplittingHostPortFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance:  Event.WebsocketUpgradeRoutine,
				Event.ClientAddress: netConn.RemoteAddr().String(),
			},
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
			Event.Context{
				Event.Circumstance:    Event.TcpSystemgeListenerAcceptRoutine,
				Event.RateLimiterType: Event.Ip,
				Event.ClientAddress:   netConn.RemoteAddr().String(),
			},
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
			Event.Context{
				Event.Circumstance:  Event.TcpSystemgeListenerAcceptRoutine,
				Event.ClientAddress: netConn.RemoteAddr().String(),
			},
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
			Event.Context{
				Event.Circumstance:  Event.TcpSystemgeListenerAcceptRoutine,
				Event.ClientAddress: netConn.RemoteAddr().String(),
			},
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
		Event.Context{
			Event.Circumstance:  Event.TcpSystemgeListenerAcceptRoutine,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientAddress: netConn.RemoteAddr().String(),
		},
	)); !event.IsInfo() {
		connection.Close()
		listener.tcpSystemgeConnectionAttemptsRejected.Add(1)
		return nil, event.GetError()
	}

	listener.tcpSystemgeConnectionAttemptsAccepted.Add(1)
	return connection, nil
}
