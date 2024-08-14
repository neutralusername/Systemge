package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *SystemgeServer) GetIncomingConnectionsList() map[string]string { // map == name:address
	server.incomingConnectionMutex.RLock()
	defer server.incomingConnectionMutex.RUnlock()
	connections := make(map[string]string, len(server.incomingConnections))
	for name, incomingConnection := range server.incomingConnections {
		connections[name] = incomingConnection.netConn.RemoteAddr().String()
	}
	return connections
}

func (server *SystemgeServer) DisconnectIncomingConnection(name string) error {
	server.incomingConnectionMutex.Lock()
	defer server.incomingConnectionMutex.Unlock()
	if server.incomingConnections[name] == nil {
		return Error.New("Connection to node \""+name+"\" does not exist", nil)
	}
	server.incomingConnections[name].netConn.Close()
	return nil
}

// adds a sync topic and its handler to systemge
// propagates the new topic of interest to all incoming connections
func (server *SystemgeServer) AddSyncTopic(topic string, handler SyncMessageHandler) error {
	server.syncMessageHandlerMutex.Lock()
	if server.syncMessageHandlers[topic] != nil {
		server.syncMessageHandlerMutex.Unlock()
		return Error.New("Sync topic already exists", nil)
	}
	server.syncMessageHandlers[topic] = handler
	server.syncMessageHandlerMutex.Unlock()
	server.topicAddSent.Add(1)
	server.incomingConnectionMutex.RLock()
	defer server.incomingConnectionMutex.RUnlock()
	for _, incomingConnection := range server.incomingConnections {
		go func() {
			if err := server.messageIncomingConnection(incomingConnection, Message.NewAsync(Message.TOPIC_ADDTOPIC, topic)); err != nil {
				if errorLogger := server.errorLogger; errorLogger != nil {
					errorLogger.Log(Error.New("Failed to send add topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error())
				}
				if mailer := server.mailer; mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send add topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error()))
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
// propagates the new topic of interest to all incoming connections
func (server *SystemgeServer) AddAsyncTopic(topic string, handler AsyncMessageHandler) error {
	server.asyncMessageHandlerMutex.Lock()
	if server.asyncMessageHandlers[topic] != nil {
		server.asyncMessageHandlerMutex.Unlock()
		return Error.New("Async topic already exists", nil)
	}
	server.asyncMessageHandlers[topic] = handler
	server.asyncMessageHandlerMutex.Unlock()
	server.topicAddSent.Add(1)
	server.incomingConnectionMutex.RLock()
	defer server.incomingConnectionMutex.RUnlock()
	for _, incomingConnection := range server.incomingConnections {
		go func() {
			if err := server.messageIncomingConnection(incomingConnection, Message.NewAsync(Message.TOPIC_ADDTOPIC, topic)); err != nil {
				if errorLogger := server.errorLogger; errorLogger != nil {
					errorLogger.Log(Error.New("Failed to send add topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error())
				}
				if mailer := server.mailer; mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send add topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error()))
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
// propagates the removal of the topic of interest to all incoming connections
func (server *SystemgeServer) RemoveSyncTopic(topic string) error {
	server.syncMessageHandlerMutex.Lock()
	if server.syncMessageHandlers[topic] == nil {
		server.syncMessageHandlerMutex.Unlock()
		return Error.New("Sync topic does not exist", nil)
	}
	delete(server.syncMessageHandlers, topic)
	server.syncMessageHandlerMutex.Unlock()
	server.topicRemoveSent.Add(1)
	server.incomingConnectionMutex.RLock()
	defer server.incomingConnectionMutex.RUnlock()
	for _, incomingConnection := range server.incomingConnections {
		go func() {
			if err := server.messageIncomingConnection(incomingConnection, Message.NewAsync(Message.TOPIC_REMOVETOPIC, topic)); err != nil {
				if errorLogger := server.errorLogger; errorLogger != nil {
					errorLogger.Log(Error.New("Failed to send remove topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error())
				}
				if mailer := server.mailer; mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send remove topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error()))
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
// propagates the removal of the topic of interest to all incoming connections
func (server *SystemgeServer) RemoveAsyncTopic(topic string) error {
	server.asyncMessageHandlerMutex.Lock()
	if server.asyncMessageHandlers[topic] == nil {
		server.asyncMessageHandlerMutex.Unlock()
		return Error.New("Async topic does not exist", nil)
	}
	delete(server.asyncMessageHandlers, topic)
	server.asyncMessageHandlerMutex.Unlock()
	server.topicRemoveSent.Add(1)
	server.incomingConnectionMutex.RLock()
	defer server.incomingConnectionMutex.RUnlock()
	for _, incomingConnection := range server.incomingConnections {
		go func() {
			if err := server.messageIncomingConnection(incomingConnection, Message.NewAsync(Message.TOPIC_REMOVETOPIC, topic)); err != nil {
				if errorLogger := server.errorLogger; errorLogger != nil {
					errorLogger.Log(Error.New("Failed to send remove topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error())
				}
				if mailer := server.mailer; mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send remove topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error()))
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
