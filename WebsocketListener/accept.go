package WebsocketListener

import (
	"errors"
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/WebsocketConnection"
)

func (listener *WebsocketListener) AcceptConnection(config *Config.WebsocketConnection, eventHandler Event.Handler) (*WebsocketConnection.WebsocketConnection, error) {

	if event := listener.onEvent(Event.NewInfo(
		Event.AcceptingConnection,
		"accepting websocketConnection",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AcceptConnection,
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	websocketConn := <-listener.connectionChannel
	if websocketConn == nil {
		listener.onEvent(Event.NewWarningNoOption(
			Event.AcceptConnectionFailed,
			"accept websocketConnection failed",
			Event.Context{
				Event.Circumstance: Event.AcceptConnection,
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
				Event.Circumstance: Event.AcceptConnection,
				Event.Address:      websocketConn.RemoteAddr().String(),
			},
		))
		listener.failed.Add(1)
		return nil, err
	}

	if listener.ipRateLimiter != nil && !listener.ipRateLimiter.RegisterConnectionAttempt(ip) {
		if event := listener.onEvent(Event.NewWarning(
			Event.RateLimited,
			"websocketConnection ip rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:    Event.AcceptConnection,
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
			"websocketConnection ip blacklisted",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.AcceptConnection,
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
			"websocketConnection ip not whitelisted",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.AcceptConnection,
				Event.Address:      websocketConn.RemoteAddr().String(),
			},
		)); !event.IsInfo() {
			listener.rejected.Add(1)
			websocketConn.Close()
			return nil, errors.New("not whitelisted")
		}
	}

	connection, err := WebsocketConnection.New(websocketConn, config, eventHandler)

	if event := listener.onEvent(Event.NewInfo(
		Event.AcceptedConnection,
		"accepted websocketConnection",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AcceptConnection,
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
