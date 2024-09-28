package SystemgeServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) AsyncMessage(topic, payload string, clientNames ...string) error {
	targetClientIds := Helpers.JsonMarshal(clientNames)
	server.statusMutex.RLock()
	server.mutex.RLock()

	if event := server.onEvent(Event.NewInfo(
		Event.SendingMultiMessage,
		"sending mutli async message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:    Event.AsyncMessage,
			Event.ClientType:      Event.SystemgeConnection,
			Event.TargetClientIds: targetClientIds,
			Event.Topic:           topic,
			Event.Payload:         payload,
		},
	)); !event.IsInfo() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
		return event.GetError()
	}

	if server.status == Status.Stopped {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStopped,
			"systemgeServer already stopped",
			Event.Context{
				Event.Circumstance:    Event.AsyncMessage,
				Event.ClientType:      Event.SystemgeConnection,
				Event.TargetClientIds: targetClientIds,
				Event.Topic:           topic,
				Event.Payload:         payload,
			},
		))
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
		return errors.New("systemgeServer already stopped")
	}

	connections := make([]SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range server.clients {
			if connection == nil {
				if event := server.onEvent(Event.NewWarning(
					Event.ClientNotAccepted,
					"client not accepted",
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance:    Event.AsyncMessage,
						Event.ClientType:      Event.SystemgeConnection,
						Event.TargetClientIds: targetClientIds,
						Event.Topic:           topic,
						Event.Payload:         payload,
					},
				)); !event.IsInfo() {
					return event.GetError()
				}
				continue
			}
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection, ok := server.clients[clientName]
			if !ok {
				if event := server.onEvent(Event.NewWarning(
					Event.ClientDoesNotExist,
					"client does not exist",
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance:    Event.AsyncMessage,
						Event.ClientType:      Event.SystemgeConnection,
						Event.TargetClientIds: targetClientIds,
						Event.Topic:           topic,
						Event.Payload:         payload,
					},
				)); !event.IsInfo() {
					return event.GetError()
				}
				continue
			}
			if connection == nil {
				if event := server.onEvent(Event.NewWarning(
					Event.ClientNotAccepted,
					"client not accepted",
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance:    Event.AsyncMessage,
						Event.ClientType:      Event.SystemgeConnection,
						Event.TargetClientIds: targetClientIds,
						Event.Topic:           topic,
						Event.Payload:         payload,
					},
				)); !event.IsInfo() {
					return event.GetError()
				}
				continue
			}
			connections = append(connections, connection)
		}
	}
	server.mutex.Unlock()
	server.statusMutex.RUnlock()

	SystemgeConnection.MultiAsyncMessage(topic, payload, connections...)

	server.onEvent(Event.NewInfoNoOption(
		Event.SentMultiMessage,
		"multi async message sent",
		Event.Context{
			Event.Circumstance:    Event.AsyncMessage,
			Event.ClientType:      Event.SystemgeConnection,
			Event.TargetClientIds: targetClientIds,
			Event.Topic:           topic,
			Event.Payload:         payload,
		},
	))
	return nil
}

func (server *SystemgeServer) SyncRequest(topic, payload string, clientNames ...string) (<-chan *Message.Message, error) {
	targetClientIds := Helpers.JsonMarshal(clientNames)
	server.statusMutex.RLock()
	server.mutex.RLock()

	if event := server.onEvent(Event.NewInfo(
		Event.SendingMultiMessage,
		"sending mutli async message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:    Event.SyncRequest,
			Event.ClientType:      Event.SystemgeConnection,
			Event.TargetClientIds: targetClientIds,
			Event.Topic:           topic,
			Event.Payload:         payload,
		},
	)); !event.IsInfo() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
		return nil, event.GetError()
	}

	if server.status == Status.Stopped {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStopped,
			"systemgeServer already stopped",
			Event.Context{
				Event.Circumstance:    Event.SyncRequest,
				Event.ClientType:      Event.SystemgeConnection,
				Event.TargetClientIds: targetClientIds,
				Event.Topic:           topic,
				Event.Payload:         payload,
			},
		))
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
		return nil, errors.New("systemgeServer already stopped")
	}

	connections := make([]SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range server.clients {
			if connection == nil {
				if event := server.onEvent(Event.NewWarning(
					Event.ClientNotAccepted,
					"client not accepted",
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance:    Event.SyncRequest,
						Event.ClientType:      Event.SystemgeConnection,
						Event.TargetClientIds: targetClientIds,
						Event.Topic:           topic,
						Event.Payload:         payload,
					},
				)); !event.IsInfo() {
					return nil, event.GetError()
				}
				continue
			}
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection, ok := server.clients[clientName]
			if !ok {
				if event := server.onEvent(Event.NewWarning(
					Event.ClientDoesNotExist,
					"client does not exist",
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance:    Event.SyncRequest,
						Event.ClientType:      Event.SystemgeConnection,
						Event.TargetClientIds: targetClientIds,
						Event.Topic:           topic,
						Event.Payload:         payload,
					},
				)); !event.IsInfo() {
					return nil, event.GetError()
				}
				continue
			}
			if connection == nil {
				if event := server.onEvent(Event.NewWarning(
					Event.ClientNotAccepted,
					"client not accepted",
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance:    Event.SyncRequest,
						Event.ClientType:      Event.SystemgeConnection,
						Event.TargetClientIds: targetClientIds,
						Event.Topic:           topic,
						Event.Payload:         payload,
					},
				)); !event.IsInfo() {
					return nil, event.GetError()
				}
				continue
			}
			connections = append(connections, connection)
		}
	}
	server.mutex.Unlock()
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
