package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (listener *WebsocketListener) Accept(config *Config.WebsocketClient, eventHandler Event.Handler) (*WebsocketClient.WebsocketClient, error) {
	if event := listener.onEvent(Event.New(
		Event.AcceptingClient,
		Event.Context{},
		Event.Continue,
		Event.Cancel,
	)); event.GetAction() == Event.Cancel {
		return nil, errors.New("accept canceled")
	}

	websocketConn := <-listener.connectionChannel //problem with stop (there might be open http requests trying to send to the channel)
	if websocketConn == nil {
		listener.onEvent(Event.New(
			Event.ReceivedNilValueFromChannel,
			Event.Context{},
			Event.Cancel,
		))
		return nil, errors.New("received nil value from connection channel")
	}

	websocketClient, err := WebsocketClient.New(config, websocketConn, eventHandler)
	if err != nil {
		listener.onEvent(Event.New(
			Event.CreateClientFailed,
			Event.Context{
				Event.Address: websocketConn.RemoteAddr().String(),
				Event.Error:   err.Error(),
			},
			Event.Cancel,
		))
		listener.clientsFailed.Add(1)
		return nil, err
	}

	if event := listener.onEvent(Event.New(
		Event.AcceptedClient,
		Event.Context{
			Event.Address: websocketConn.RemoteAddr().String(),
		},
		Event.Continue,
		Event.Cancel,
	)); event.GetAction() == Event.Cancel {
		websocketClient.Close()
		listener.clientsRejected.Add(1)
		return nil, errors.New("client rejected")
	}

	listener.clientsAccepted.Add(1)
	return websocketClient, nil
}

/*
	ip, _, err := net.SplitHostPort(websocketConn.RemoteAddr().String())
	if err != nil {
		listener.onEvent(Event.NewWarningNoOption(
			Event.SplittingHostPortFailed,
			err.Error(),
			Event.Context{
				Event.Address: websocketConn.RemoteAddr().String(),
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
				Event.Address: websocketConn.RemoteAddr().String(),
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
				Event.Address: websocketConn.RemoteAddr().String(),
			},
		)); !event.IsInfo() {
			listener.clientsRejected.Add(1)
			websocketConn.Close()
			return nil, errors.New("not whitelisted")
		}
	}

*/
