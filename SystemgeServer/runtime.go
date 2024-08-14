package SystemgeServer

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *SystemgeServer) GetClientConnectionsList() map[string]string { // map == name:address
	server.clientConnectionMutex.RLock()
	defer server.clientConnectionMutex.RUnlock()
	connections := make(map[string]string, len(server.clientConnections))
	for name, clientConnection := range server.clientConnections {
		connections[name] = clientConnection.netConn.RemoteAddr().String()
	}
	return connections
}

func (server *SystemgeServer) DisconnectClientConnection(name string) error {
	server.clientConnectionMutex.Lock()
	defer server.clientConnectionMutex.Unlock()
	if server.clientConnections[name] == nil {
		return Error.New("Connection with name \""+name+"\" does not exist", nil)
	}
	server.clientConnections[name].netConn.Close()
	return nil
}

// adds a sync topic and its handler to systemge
// propagates the new topic of interest to all client connections
func (server *SystemgeServer) AddSyncTopic(topic string, handler SyncMessageHandler) error {
	server.syncMessageHandlerMutex.Lock()
	if server.syncMessageHandlers[topic] != nil {
		server.syncMessageHandlerMutex.Unlock()
		return Error.New("Sync topic already exists", nil)
	}
	server.syncMessageHandlers[topic] = handler
	server.syncMessageHandlerMutex.Unlock()
	server.topicAddSent.Add(1)
	server.clientConnectionMutex.RLock()
	defer server.clientConnectionMutex.RUnlock()
	for _, clientConnection := range server.clientConnections {
		go func() {
			if err := server.messageClientConnection(clientConnection, Message.NewAsync(Message.TOPIC_ADDTOPIC, topic)); err != nil {
				if errorLogger := server.errorLogger; errorLogger != nil {
					errorLogger.Log(Error.New("Failed to send add topic message to client connection \""+clientConnection.name+"\"", err).Error())
				}
				if mailer := server.mailer; mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send add topic message to client connection \""+clientConnection.name+"\"", err).Error()))
					if err != nil {
						if errorLogger := server.errorLogger; errorLogger != nil {
							errorLogger.Log(Error.New("Failed sending mail", err).Error())
						}
					}
				}
			}
		}()
	}
	return nil
}

// adds an async topic and its handler to systemge
// propagates the new topic of interest to all client connections
func (server *SystemgeServer) AddAsyncTopic(topic string, handler AsyncMessageHandler) error {
	server.asyncMessageHandlerMutex.Lock()
	if server.asyncMessageHandlers[topic] != nil {
		server.asyncMessageHandlerMutex.Unlock()
		return Error.New("Async topic already exists", nil)
	}
	server.asyncMessageHandlers[topic] = handler
	server.asyncMessageHandlerMutex.Unlock()
	server.topicAddSent.Add(1)
	server.clientConnectionMutex.RLock()
	defer server.clientConnectionMutex.RUnlock()
	for _, clientConnection := range server.clientConnections {
		go func() {
			if err := server.messageClientConnection(clientConnection, Message.NewAsync(Message.TOPIC_ADDTOPIC, topic)); err != nil {
				if errorLogger := server.errorLogger; errorLogger != nil {
					errorLogger.Log(Error.New("Failed to send add topic message to client connection \""+clientConnection.name+"\"", err).Error())
				}
				if mailer := server.mailer; mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send add topic message to client connection \""+clientConnection.name+"\"", err).Error()))
					if err != nil {
						if errorLogger := server.errorLogger; errorLogger != nil {
							errorLogger.Log(Error.New("Failed sending mail", err).Error())
						}
					}
				}
			}
		}()
	}
	return nil
}

// removes a sync topic and its handler from systemge
// propagates the removal of the topic of interest to all client connections
func (server *SystemgeServer) RemoveSyncTopic(topic string) error {
	server.syncMessageHandlerMutex.Lock()
	if server.syncMessageHandlers[topic] == nil {
		server.syncMessageHandlerMutex.Unlock()
		return Error.New("Sync topic does not exist", nil)
	}
	delete(server.syncMessageHandlers, topic)
	server.syncMessageHandlerMutex.Unlock()
	server.topicRemoveSent.Add(1)
	server.clientConnectionMutex.RLock()
	defer server.clientConnectionMutex.RUnlock()
	for _, clientConnection := range server.clientConnections {
		go func() {
			if err := server.messageClientConnection(clientConnection, Message.NewAsync(Message.TOPIC_REMOVETOPIC, topic)); err != nil {
				if errorLogger := server.errorLogger; errorLogger != nil {
					errorLogger.Log(Error.New("Failed to send remove topic message to client connection \""+clientConnection.name+"\"", err).Error())
				}
				if mailer := server.mailer; mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send remove topic message to client connection \""+clientConnection.name+"\"", err).Error()))
					if err != nil {
						if errorLogger := server.errorLogger; errorLogger != nil {
							errorLogger.Log(Error.New("Failed sending mail", err).Error())
						}
					}
				}
			}
		}()
	}
	return nil
}

// removes an async topic and its handler from systemge
// propagates the removal of the topic of interest to all client connections
func (server *SystemgeServer) RemoveAsyncTopic(topic string) error {
	server.asyncMessageHandlerMutex.Lock()
	if server.asyncMessageHandlers[topic] == nil {
		server.asyncMessageHandlerMutex.Unlock()
		return Error.New("Async topic does not exist", nil)
	}
	delete(server.asyncMessageHandlers, topic)
	server.asyncMessageHandlerMutex.Unlock()
	server.topicRemoveSent.Add(1)
	server.clientConnectionMutex.RLock()
	defer server.clientConnectionMutex.RUnlock()
	for _, clientConnection := range server.clientConnections {
		go func() {
			if err := server.messageClientConnection(clientConnection, Message.NewAsync(Message.TOPIC_REMOVETOPIC, topic)); err != nil {
				if errorLogger := server.errorLogger; errorLogger != nil {
					errorLogger.Log(Error.New("Failed to send remove topic message to client connection \""+clientConnection.name+"\"", err).Error())
				}
				if mailer := server.mailer; mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send remove topic message to client connection \""+clientConnection.name+"\"", err).Error()))
					if err != nil {
						if errorLogger := server.errorLogger; errorLogger != nil {
							errorLogger.Log(Error.New("Failed sending mail", err).Error())
						}
					}
				}
			}
		}()
	}
	return nil
}

// AddToSystemgeBlacklist adds an address to the systemge blacklist.
func (server *SystemgeServer) AddToSystemgeBlacklist(address string) {
	server.tcpServer.GetBlacklist().Add(address)
}

// RemoveFromSystemgeBlacklist removes an address from the systemge blacklist.
func (server *SystemgeServer) RemoveFromSystemgeBlacklist(address string) {
	server.tcpServer.GetBlacklist().Remove(address)
}

// GetSystemgeBlacklist returns a slice of addresses in the systemge blacklist.
func (server *SystemgeServer) GetSystemgeBlacklist() []string {
	return server.tcpServer.GetBlacklist().GetElements()
}

// AddToSystemgeWhitelist adds an address to the systemge whitelist.
func (server *SystemgeServer) AddToSystemgeWhitelist(address string) {
	server.tcpServer.GetWhitelist().Add(address)
}

// RemoveFromSystemgeWhitelist removes an address from the systemge whitelist.
func (server *SystemgeServer) RemoveFromSystemgeWhitelist(address string) {
	server.tcpServer.GetWhitelist().Remove(address)
}

// GetSystemgeWhitelist returns a slice of addresses in the systemge whitelist.
func (server *SystemgeServer) GetSystemgeWhitelist() []string {
	return server.tcpServer.GetWhitelist().GetElements()
}
