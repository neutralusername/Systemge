package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

func (client *SystemgeClient) AsyncMessage(topic, payload string, clientNames ...string) error {
	client.statusMutex.RLock()
	if client.status == Status.Stopped {
		client.statusMutex.RUnlock()
		return Event.New("Client stopped", nil)
	}
	client.mutex.Lock()
	connections := make([]SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range client.nameConnections {
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection := client.nameConnections[clientName]
			if connection == nil {
				if client.errorLogger != nil {
					client.errorLogger.Log(Event.New("Client \""+clientName+"\" not found for async message", nil).Error())
				}
				if client.mailer != nil {
					err := client.mailer.Send(Tools.NewMail(nil, "error", Event.New("Client \""+clientName+"\" not found for async message", nil).Error()))
					if err != nil {
						if client.errorLogger != nil {
							client.errorLogger.Log(Event.New("failed sending mail", err).Error())
						}
					}
				}
				continue
			}
			connections = append(connections, connection)
		}
	}
	client.mutex.Unlock()
	client.statusMutex.RUnlock()

	errorChannel := SystemgeConnection.MultiAsyncMessage(topic, payload, connections...)
	go func() {
		for err := range errorChannel {
			if client.errorLogger != nil {
				client.errorLogger.Log(err.Error())
			}
			if client.mailer != nil {
				err := client.mailer.Send(Tools.NewMail(nil, "error", err.Error()))
				if err != nil {
					if client.errorLogger != nil {
						client.errorLogger.Log(Event.New("failed sending mail", err).Error())
					}
				}
			}
		}
	}()
	return nil
}

func (client *SystemgeClient) SyncRequestBlocking(topic, payload string, clientNames ...string) ([]*Message.Message, error) {
	client.statusMutex.RLock()
	if client.status == Status.Stopped {
		client.statusMutex.RUnlock()
		return nil, Event.New("Client stopped", nil)
	}
	client.mutex.Lock()
	connections := make([]SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range client.nameConnections {
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection := client.nameConnections[clientName]
			if connection == nil {
				if client.errorLogger != nil {
					client.errorLogger.Log(Event.New("Client \""+clientName+"\" not found for sync request", nil).Error())
				}
				if client.mailer != nil {
					err := client.mailer.Send(Tools.NewMail(nil, "error", Event.New("Client \""+clientName+"\" not found for sync request", nil).Error()))
					if err != nil {
						if client.errorLogger != nil {
							client.errorLogger.Log(Event.New("failed sending mail", err).Error())
						}
					}
				}
				continue
			}
			connections = append(connections, connection)
		}
	}
	client.mutex.Unlock()
	client.statusMutex.RUnlock()

	responseChannel, errorChannel := SystemgeConnection.MultiSyncRequest(topic, payload, connections...)
	responses := make([]*Message.Message, 0)
	for response := range responseChannel {
		responses = append(responses, response)
	}
	for err := range errorChannel {
		if client.errorLogger != nil {
			client.errorLogger.Log(err.Error())
			if client.mailer != nil {
				err := client.mailer.Send(Tools.NewMail(nil, "error", err.Error()))
				if err != nil {
					if client.errorLogger != nil {
						client.errorLogger.Log(Event.New("failed sending mail", err).Error())
					}
				}
			}
		}
	}
	return responses, nil
}

func (client *SystemgeClient) SyncRequest(topic, payload string, clientNames ...string) (<-chan *Message.Message, error) {
	client.statusMutex.RLock()
	if client.status == Status.Stopped {
		client.statusMutex.RUnlock()
		return nil, nil
	}
	client.mutex.Lock()
	connections := make([]SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range client.nameConnections {
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection := client.nameConnections[clientName]
			if connection == nil {
				if client.errorLogger != nil {
					client.errorLogger.Log(Event.New("Client \""+clientName+"\" not found for sync request", nil).Error())
				}
				if client.mailer != nil {
					err := client.mailer.Send(Tools.NewMail(nil, "error", Event.New("Client \""+clientName+"\" not found for sync request", nil).Error()))
					if err != nil {
						if client.errorLogger != nil {
							client.errorLogger.Log(Event.New("failed sending mail", err).Error())
						}
					}
				}
				continue
			}
			connections = append(connections, connection)
		}
	}
	client.mutex.Unlock()
	client.statusMutex.RUnlock()

	responseChannel, errorChannel := SystemgeConnection.MultiSyncRequest(topic, payload, connections...)
	go func() {
		for err := range errorChannel {
			if client.errorLogger != nil {
				client.errorLogger.Log(err.Error())
			}
			if client.mailer != nil {
				err := client.mailer.Send(Tools.NewMail(nil, "error", err.Error()))
				if err != nil {
					if client.errorLogger != nil {
						client.errorLogger.Log(Event.New("failed sending mail", err).Error())
					}
				}
			}
		}
	}()
	return responseChannel, nil
}
