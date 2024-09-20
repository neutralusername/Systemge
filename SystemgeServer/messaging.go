package SystemgeServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *SystemgeServer) AsyncMessage(topic, payload string, clientNames ...string) error {
	server.statusMutex.RLock()
	if server.status == Status.STOPPED {
		server.statusMutex.RUnlock()
		return Event.New("Server stopped", nil)
	}
	server.mutex.Lock()
	connections := make([]SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range server.clients {
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection := server.clients[clientName]
			if connection == nil {
				if server.errorLogger != nil {
					server.errorLogger.Log(Event.New("Client \""+clientName+"\" not found", nil).Error())
				}
				if server.mailer != nil {
					err := server.mailer.Send(Tools.NewMail(nil, "error", Event.New("Client \""+clientName+"\" not found", nil).Error()))
					if err != nil {
						if server.errorLogger != nil {
							server.errorLogger.Log(Event.New("failed sending mail", err).Error())
						}
					}
				}
				continue
			}
			connections = append(connections, connection)
		}
	}
	server.mutex.Unlock()
	server.statusMutex.RUnlock()

	errorChannel := SystemgeConnection.MultiAsyncMessage(topic, payload, connections...)
	go func() {
		for err := range errorChannel {
			if server.errorLogger != nil {
				server.errorLogger.Log(err.Error())
			}
			if server.mailer != nil {
				err := server.mailer.Send(Tools.NewMail(nil, "error", err.Error()))
				if err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Event.New("failed sending mail", err).Error())
					}
				}
			}
		}
	}()
	return nil
}

func (server *SystemgeServer) SyncRequest(topic, payload string, clientNames ...string) (<-chan *Message.Message, error) {
	server.statusMutex.RLock()
	if server.status == Status.STOPPED {
		server.statusMutex.RUnlock()
		return nil, Event.New("Server stopped", nil)
	}
	server.mutex.Lock()
	connections := make([]SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range server.clients {
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection := server.clients[clientName]
			if connection == nil {
				if server.errorLogger != nil {
					server.errorLogger.Log(Event.New("Client \""+clientName+"\" not found", nil).Error())
				}
				if server.mailer != nil {
					err := server.mailer.Send(Tools.NewMail(nil, "error", Event.New("Client \""+clientName+"\" not found", nil).Error()))
					if err != nil {
						if server.errorLogger != nil {
							server.errorLogger.Log(Event.New("failed sending mail", err).Error())
						}
					}
				}
				continue
			}
			connections = append(connections, connection)
		}
	}
	server.mutex.Unlock()
	server.statusMutex.RUnlock()

	responseChannel, errorChannel := SystemgeConnection.MultiSyncRequest(topic, payload, connections...)
	go func() {
		for err := range errorChannel {
			if server.errorLogger != nil {
				server.errorLogger.Log(err.Error())
			}
			if server.mailer != nil {
				err := server.mailer.Send(Tools.NewMail(nil, "error", err.Error()))
				if err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Event.New("failed sending mail", err).Error())
					}
				}
			}
		}
	}()
	return responseChannel, nil
}

func (server *SystemgeServer) SyncRequestBlocking(topic, payload string, clientNames ...string) ([]*Message.Message, error) {
	server.statusMutex.RLock()
	if server.status == Status.STOPPED {
		server.statusMutex.RUnlock()
		return nil, Event.New("Server stopped", nil)
	}
	server.mutex.Lock()
	connections := make([]SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range server.clients {
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection := server.clients[clientName]
			if connection == nil {
				if server.errorLogger != nil {
					server.errorLogger.Log(Event.New("Client \""+clientName+"\" not found", nil).Error())
				}
				if server.mailer != nil {
					err := server.mailer.Send(Tools.NewMail(nil, "error", Event.New("Client \""+clientName+"\" not found", nil).Error()))
					if err != nil {
						if server.errorLogger != nil {
							server.errorLogger.Log(Event.New("failed sending mail", err).Error())
						}
					}
				}
				continue
			}
			connections = append(connections, connection)
		}
	}
	server.mutex.Unlock()
	server.statusMutex.RUnlock()

	responseChannel, errorChannel := SystemgeConnection.MultiSyncRequest(topic, payload, connections...)
	responses := make([]*Message.Message, 0)
	for response := range responseChannel {
		responses = append(responses, response)
	}
	for err := range errorChannel {
		if server.errorLogger != nil {
			server.errorLogger.Log(err.Error())
		}
		if server.mailer != nil {
			err := server.mailer.Send(Tools.NewMail(nil, "error", err.Error()))
			if err != nil {
				if server.errorLogger != nil {
					server.errorLogger.Log(Event.New("failed sending mail", err).Error())
				}
			}
		}
	}
	return responses, nil
}
