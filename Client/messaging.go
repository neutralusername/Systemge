package Client

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/helpers"
	"github.com/neutralusername/Systemge/status"
)

func (client *Client) AsyncMessage(topic, payload string, clientNames ...string) error {
	targetClientIds := helpers.JsonMarshal(clientNames)
	client.statusMutex.RLock()
	client.mutex.RLock()

	if event := client.onEvent(Event.NewInfo(
		Event.SendingMultiMessage,
		"sending mutli async message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AsyncMessage,
			Event.Targets:      targetClientIds,
			Event.Topic:        topic,
			Event.Payload:      payload,
		},
	)); !event.IsInfo() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
		return event.GetError()
	}

	if client.status == status.Stopped {
		client.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStoped,
			"systemgeClient already stopped",
			Event.Context{
				Event.Circumstance: Event.AsyncMessage,
				Event.Targets:      targetClientIds,
				Event.Topic:        topic,
				Event.Payload:      payload,
			},
		))
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
		return errors.New("systemgeClient already stopped")
	}

	connections := make([]SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range client.nameConnections {
			if connection == nil {
				if event := client.onEvent(Event.NewWarning(
					Event.ClientNotAccepted,
					"client not accepted",
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance: Event.AsyncMessage,
						Event.Targets:      targetClientIds,
						Event.Topic:        topic,
						Event.Payload:      payload,
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
			connection, ok := client.nameConnections[clientName]
			if !ok {
				if event := client.onEvent(Event.NewWarning(
					Event.SessionDoesNotExist,
					"client does not exist",
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance: Event.AsyncMessage,
						Event.Targets:      targetClientIds,
						Event.Topic:        topic,
						Event.Payload:      payload,
					},
				)); !event.IsInfo() {
					return event.GetError()
				}
				continue
			}
			if connection == nil {
				if event := client.onEvent(Event.NewWarning(
					Event.ClientNotAccepted,
					"client not accepted",
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance: Event.AsyncMessage,
						Event.Targets:      targetClientIds,
						Event.Topic:        topic,
						Event.Payload:      payload,
					},
				)); !event.IsInfo() {
					return event.GetError()
				}
				continue
			}
			connections = append(connections, connection)
		}
	}
	client.mutex.Unlock()
	client.statusMutex.RUnlock()

	SystemgeConnection.MultiWrite(topic, payload, connections...)

	client.onEvent(Event.NewInfoNoOption(
		Event.SentMultiMessage,
		"multi async message sent",
		Event.Context{
			Event.Circumstance: Event.AsyncMessage,
			Event.Targets:      targetClientIds,
			Event.Topic:        topic,
			Event.Payload:      payload,
		},
	))
	return nil
}

func (client *Client) SyncRequest(topic, payload string, clientNames ...string) (<-chan *Message.Message, error) {
	targetClientIds := helpers.JsonMarshal(clientNames)
	client.statusMutex.RLock()
	client.mutex.RLock()

	if event := client.onEvent(Event.NewInfo(
		Event.SendingMultiMessage,
		"sending mutli sync message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.SyncRequest,
			Event.Targets:      targetClientIds,
			Event.Topic:        topic,
			Event.Payload:      payload,
		},
	)); !event.IsInfo() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
		return nil, event.GetError()
	}

	if client.status == status.Stopped {
		client.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStoped,
			"systemgeClient already stopped",
			Event.Context{
				Event.Circumstance: Event.SyncRequest,
				Event.Targets:      targetClientIds,
				Event.Topic:        topic,
				Event.Payload:      payload,
			},
		))
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
		return nil, errors.New("systemgeClient already stopped")
	}

	connections := make([]SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range client.nameConnections {
			if connection == nil {
				if event := client.onEvent(Event.NewWarning(
					Event.ClientNotAccepted,
					"client not accepted",
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance: Event.SyncRequest,
						Event.Targets:      targetClientIds,
						Event.Topic:        topic,
						Event.Payload:      payload,
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
			connection, ok := client.nameConnections[clientName]
			if !ok {
				if event := client.onEvent(Event.NewWarning(
					Event.SessionDoesNotExist,
					"client does not exist",
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance: Event.SyncRequest,
						Event.Targets:      targetClientIds,
						Event.Topic:        topic,
						Event.Payload:      payload,
					},
				)); !event.IsInfo() {
					return nil, event.GetError()
				}
				continue
			}
			if connection == nil {
				if event := client.onEvent(Event.NewWarning(
					Event.SessionNotAccepted,
					"client not accepted",
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance: Event.SyncRequest,
						Event.Targets:      targetClientIds,
						Event.Topic:        topic,
						Event.Payload:      payload,
					},
				)); !event.IsInfo() {
					return nil, event.GetError()
				}
				continue
			}
			connections = append(connections, connection)
		}
	}
	client.mutex.Unlock()
	client.statusMutex.RUnlock()

	return SystemgeConnection.MultiSyncRequest(topic, payload, connections...), nil
}

func (client *Client) SyncRequestBlocking(topic, payload string, clientNames ...string) ([]*Message.Message, error) {
	responseChannel, err := client.SyncRequest(topic, payload, clientNames...)
	if err != nil {
		return nil, err
	}
	responses := []*Message.Message{}
	for response := range responseChannel {
		responses = append(responses, response)
	}
	return responses, nil
}
