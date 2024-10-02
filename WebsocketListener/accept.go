package WebsocketListener

import (
	"errors"
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/WebsocketConnection"
)

func (listener *WebsocketListener) AcceptConnection(connectionConfig *Config.TcpSystemgeConnection, eventHandler Event.Handler) (*WebsocketConnection.WebsocketConnection, error) {

	if event := listener.onEvent(Event.NewInfo(
		Event.AcceptingTcpSystemgeConnection,
		"accepting TcpSystemgeConnection",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AcceptTcpSystemgeConnection,
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	websocketConn := <-listener.connectionChannel
	if websocketConn == nil {
		listener.onEvent(Event.NewWarningNoOption(
			Event.AcceptTcpListenerFailed,
			"accepting TcpSystemgeConnection",
			Event.Context{
				Event.Circumstance: Event.AcceptTcpSystemgeConnection,
			},
		))
		listener.failed.Add(1)
	}

	ip, _, err := net.SplitHostPort(websocketConn.RemoteAddr().String())
	if err != nil {
		listener.onEvent(Event.NewWarningNoOption(
			Event.SplittingHostPortFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.AcceptTcpSystemgeConnection,
				Event.Address:      websocketConn.RemoteAddr().String(),
			},
		))
		listener.failed.Add(1)
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
				Event.Circumstance:    Event.AcceptTcpSystemgeConnection,
				Event.RateLimiterType: Event.Ip,
				Event.Address:         websocketConn.RemoteAddr().String(),
			},
		)); !event.IsInfo() {
			listener.rejected.Add(1)
			websocketConn.Close()
			return nil, errors.New("rate limit exceeded")
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
				Event.Circumstance: Event.AcceptTcpSystemgeConnection,
				Event.Address:      websocketConn.RemoteAddr().String(),
			},
		)); !event.IsInfo() {
			listener.rejected.Add(1)
			websocketConn.Close()
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
				Event.Circumstance: Event.AcceptTcpSystemgeConnection,
				Event.Address:      websocketConn.RemoteAddr().String(),
			},
		)); !event.IsInfo() {
			listener.rejected.Add(1)
			websocketConn.Close()
			return nil, errors.New("not whitelisted")
		}
	}

	connection, err := listener.serverHandshake(connectionConfig, netConn, eventHandler)
	if err != nil {
		listener.onEvent(Event.NewWarningNoOption(
			Event.ServerHandshakeFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.AcceptTcpSystemgeConnection,
				Event.Address:      websocketConn.RemoteAddr().String(),
			},
		))
		listener.rejected.Add(1)
		websocketConn.Close()
		return nil, err
	}

	if event := listener.onEvent(Event.NewInfo(
		Event.AcceptedTcpSystemgeConnection,
		"accepted TcpSystemgeConnection",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AcceptTcpSystemgeConnection,
			Event.Address:      websocketConn.RemoteAddr().String(),
		},
	)); !event.IsInfo() {
		connection.Close()
		listener.rejected.Add(1)
		return nil, event.GetError()
	}

	listener.accepted.Add(1)
	return connection, nil
}
