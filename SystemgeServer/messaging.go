package SystemgeServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) AsyncMessage(topic, payload string, identities ...string) error {
	targetIdentities := Helpers.JsonMarshal(identities)
	server.statusMutex.RLock()

	if event := server.onEvent(Event.NewInfo(
		Event.SendingMultiMessage,
		"sending mutli async message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AsyncMessage,
			Event.Targets:      targetIdentities,
			Event.Topic:        topic,
			Event.Payload:      payload,
		},
	)); !event.IsInfo() {
		server.statusMutex.RUnlock()
		return event.GetError()
	}

	if server.status == Status.Stopped {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStoped,
			"systemgeServer already stopped",
			Event.Context{
				Event.Circumstance: Event.AsyncMessage,
				Event.Targets:      targetIdentities,
				Event.Topic:        topic,
				Event.Payload:      payload,
			},
		))
		server.statusMutex.RUnlock()
		return errors.New("systemgeServer already stopped")
	}

	connections := make([]SystemgeConnection.SystemgeConnection, 0)
	if len(identities) == 0 {
		for _, identity := range server.sessionManager.GetIdentities() {
			for _, session := range server.sessionManager.GetSessions(identity) {
				if !session.IsAccepted() {
					event := server.onEvent(Event.NewWarning(
						Event.SessionNotAccepted,
						"session not accepted",
						Event.Cancel,
						Event.Skip,
						Event.Continue,
						Event.Context{
							Event.Circumstance: Event.AsyncMessage,
							Event.Targets:      targetIdentities,
							Event.Topic:        topic,
							Event.Payload:      payload,
						},
					))
					if event.IsError() {
						server.statusMutex.RUnlock()
						return event.GetError()
					}
					if event.IsWarning() {
						continue
					}
				}
				connection, ok := session.Get("connection")
				if !ok {
					continue
				}
				connections = append(connections, connection.(SystemgeConnection.SystemgeConnection))
			}
		}
	} else {
		for _, identity := range identities {
			sessions := server.sessionManager.GetSessions(identity)
			if len(sessions) == 0 {
				if event := server.onEvent(Event.NewWarning(
					Event.SessionDoesNotExist,
					"client does not exist",
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance: Event.SyncRequest,
						Event.Targets:      targetIdentities,
						Event.Topic:        topic,
						Event.Payload:      payload,
					},
				)); !event.IsInfo() {
					return event.GetError()
				}
			}
			for _, session := range sessions {
				if !session.IsAccepted() {
					event := server.onEvent(Event.NewWarning(
						Event.SessionNotAccepted,
						"session not accepted",
						Event.Cancel,
						Event.Skip,
						Event.Continue,
						Event.Context{
							Event.Circumstance: Event.AsyncMessage,
							Event.Targets:      targetIdentities,
							Event.Topic:        topic,
							Event.Payload:      payload,
						},
					))
					if event.IsError() {
						server.statusMutex.RUnlock()
						return event.GetError()
					}
					if event.IsWarning() {
						continue
					}
				}
				connection, ok := session.Get("connection")
				if !ok {
					continue
				}
				connections = append(connections, connection.(SystemgeConnection.SystemgeConnection))
			}
		}
	}
	server.statusMutex.RUnlock()

	SystemgeConnection.MultiAsyncMessage(topic, payload, connections...)

	server.onEvent(Event.NewInfoNoOption(
		Event.SentMultiMessage,
		"multi async message sent",
		Event.Context{
			Event.Circumstance: Event.AsyncMessage,
			Event.Targets:      targetIdentities,
			Event.Topic:        topic,
			Event.Payload:      payload,
		},
	))
	return nil
}

func (server *SystemgeServer) SyncRequest(topic, payload string, identities ...string) (<-chan *Message.Message, error) {
	targetIdentities := Helpers.JsonMarshal(identities)
	server.statusMutex.RLock()

	if event := server.onEvent(Event.NewInfo(
		Event.SendingMultiMessage,
		"sending mutli sync message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.SyncRequest,
			Event.Targets:      targetIdentities,
			Event.Topic:        topic,
			Event.Payload:      payload,
		},
	)); !event.IsInfo() {
		server.statusMutex.RUnlock()
		return nil, event.GetError()
	}

	if server.status == Status.Stopped {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStoped,
			"systemgeServer already stopped",
			Event.Context{
				Event.Circumstance: Event.SyncRequest,
				Event.Targets:      targetIdentities,
				Event.Topic:        topic,
				Event.Payload:      payload,
			},
		))
		server.statusMutex.RUnlock()
		return nil, errors.New("systemgeServer already stopped")
	}

	connections := make([]SystemgeConnection.SystemgeConnection, 0)
	if len(identities) == 0 {
		for _, identity := range server.sessionManager.GetIdentities() {
			for _, session := range server.sessionManager.GetSessions(identity) {
				if !session.IsAccepted() {
					event := server.onEvent(Event.NewWarning(
						Event.SessionNotAccepted,
						"session not accepted",
						Event.Cancel,
						Event.Skip,
						Event.Continue,
						Event.Context{
							Event.Circumstance: Event.AsyncMessage,
							Event.Targets:      targetIdentities,
							Event.Topic:        topic,
							Event.Payload:      payload,
						},
					))
					if event.IsError() {
						server.statusMutex.RUnlock()
						return nil, event.GetError()
					}
					if event.IsWarning() {
						continue
					}
				}
				connection, ok := session.Get("connection")
				if !ok {
					continue
				}
				connections = append(connections, connection.(SystemgeConnection.SystemgeConnection))
			}
		}
	} else {
		for _, identity := range identities {
			sessions := server.sessionManager.GetSessions(identity)
			if len(sessions) == 0 {
				if event := server.onEvent(Event.NewWarning(
					Event.SessionDoesNotExist,
					"client does not exist",
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance: Event.SyncRequest,
						Event.Targets:      targetIdentities,
						Event.Topic:        topic,
						Event.Payload:      payload,
					},
				)); !event.IsInfo() {
					return nil, event.GetError()
				}
			}
			for _, session := range sessions {
				if !session.IsAccepted() {
					event := server.onEvent(Event.NewWarning(
						Event.SessionNotAccepted,
						"session not accepted",
						Event.Cancel,
						Event.Skip,
						Event.Continue,
						Event.Context{
							Event.Circumstance: Event.AsyncMessage,
							Event.Targets:      targetIdentities,
							Event.Topic:        topic,
							Event.Payload:      payload,
						},
					))
					if event.IsError() {
						server.statusMutex.RUnlock()
						return nil, event.GetError()
					}
					if event.IsWarning() {
						continue
					}
				}
				connection, ok := session.Get("connection")
				if !ok {
					continue
				}
				connections = append(connections, connection.(SystemgeConnection.SystemgeConnection))
			}
		}
	}
	server.statusMutex.RUnlock()

	return SystemgeConnection.MultiSyncRequest(topic, payload, connections...), nil
}

func (server *SystemgeServer) SyncRequestBlocking(topic, payload string, clientNames ...string) ([]*Message.Message, error) {
	responseChannel, err := server.SyncRequest(topic, payload, clientNames...)
	if err != nil {
		return nil, err
	}
	responses := []*Message.Message{}
	for response := range responseChannel {
		responses = append(responses, response)
	}
	return responses, nil
}
