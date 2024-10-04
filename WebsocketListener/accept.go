package WebsocketListener

import (
	"errors"
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (listener *WebsocketListener) AcceptClient(config *Config.WebsocketClient, eventHandler Event.Handler) (*WebsocketClient.WebsocketClient, error) {
	if event := listener.onEvent(Event.NewInfo(
		Event.AcceptingClient,
		"accepting websocketClient",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AcceptClient,
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	websocketConn := <-listener.connectionChannel
	if websocketConn == nil {
		listener.onEvent(Event.NewWarningNoOption(
			Event.ReceivedNilValueFromChannel,
			"received nil value from connection channel",
			Event.Context{
				Event.Circumstance: Event.AcceptClient,
			},
		))
		listener.clientsFailed.Add(1)
		return nil, errors.New("received nil value from connection channel")
	}

	ip, _, err := net.SplitHostPort(websocketConn.RemoteAddr().String())
	if err != nil {
		listener.onEvent(Event.NewWarningNoOption(
			Event.SplittingHostPortFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.AcceptClient,
				Event.Address:      websocketConn.RemoteAddr().String(),
			},
		))
		listener.clientsFailed.Add(1)
		return nil, err
	}

	if listener.ipRateLimiter != nil && !listener.ipRateLimiter.RegisterConnectionAttempt(ip) {
		if event := listener.onEvent(Event.NewWarning(
			Event.RateLimited,
			"websocketClient ip rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:    Event.AcceptClient,
				Event.RateLimiterType: Event.Ip,
				Event.Address:         websocketConn.RemoteAddr().String(),
			},
		)); !event.IsInfo() {
			listener.clientsRejected.Add(1)
			websocketConn.Close()
			return nil, errors.New("rate limit exceeded")
		}
	}

	if listener.blacklist != nil && listener.blacklist.Contains(ip) {
		if event := listener.onEvent(Event.NewWarning(
			Event.Blacklisted,
			"websocketClient ip blacklisted",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.AcceptClient,
				Event.Address:      websocketConn.RemoteAddr().String(),
			},
		)); !event.IsInfo() {
			listener.clientsRejected.Add(1)
			websocketConn.Close()
			return nil, errors.New("Blacklisted")
		}
	}

	if listener.whitelist != nil && listener.whitelist.ElementCount() > 0 && !listener.whitelist.Contains(ip) {
		if event := listener.onEvent(Event.NewWarning(
			Event.NotWhitelisted,
			"websocketClient ip not whitelisted",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.AcceptClient,
				Event.Address:      websocketConn.RemoteAddr().String(),
			},
		)); !event.IsInfo() {
			listener.clientsRejected.Add(1)
			websocketConn.Close()
			return nil, errors.New("not whitelisted")
		}
	}

	websocketClient, err := WebsocketClient.New(clientName, config, websocketConn, eventHandler)

	if event := listener.onEvent(Event.NewInfo(
		Event.AcceptedClient,
		"accepted websocketClient",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AcceptClient,
			Event.Address:      websocketConn.RemoteAddr().String(),
		},
	)); !event.IsInfo() {
		websocketClient.Close()
		listener.clientsRejected.Add(1)
		return nil, event.GetError()
	}

	listener.clientsAccepted.Add(1)
	return websocketClient, nil
}
