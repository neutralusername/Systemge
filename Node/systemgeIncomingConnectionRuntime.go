package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (node *Node) GetIncomingConnectionsList() map[string]string { // map == name:address
	if systemge := node.systemge; systemge != nil {
		systemge.incomingConnectionMutex.RLock()
		defer systemge.incomingConnectionMutex.RUnlock()
		connections := make(map[string]string, len(systemge.incomingConnections))
		for name, incomingConnection := range systemge.incomingConnections {
			connections[name] = incomingConnection.netConn.RemoteAddr().String()
		}
		return connections
	}
	return nil
}

func (node *Node) DisconnectIncomingConnection(name string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.incomingConnectionMutex.Lock()
		defer systemge.incomingConnectionMutex.Unlock()
		if systemge.incomingConnections[name] == nil {
			return Error.New("Connection to node \""+name+"\" does not exist", nil)
		}
		systemge.incomingConnections[name].netConn.Close()
		return nil
	}
	return Error.New("Systemge is nil", nil)
}

// adds a sync topic and its handler to systemge
// propagates the new topic of interest to all incoming connections
func (node *Node) AddSyncTopic(topic string, handler SyncMessageHandler) error {
	if systemge := node.systemge; systemge != nil {
		systemge.syncMessageHandlerMutex.Lock()
		if systemge.syncMessageHandlers[topic] != nil {
			systemge.syncMessageHandlerMutex.Unlock()
			return Error.New("Sync topic already exists", nil)
		}
		systemge.syncMessageHandlers[topic] = handler
		systemge.syncMessageHandlerMutex.Unlock()
		systemge.sentAddTopic.Add(1)
		systemge.incomingConnectionMutex.RLock()
		defer systemge.incomingConnectionMutex.RUnlock()
		for _, incomingConnection := range systemge.incomingConnections {
			go func() {
				if err := systemge.messageIncomingConnection(incomingConnection, Message.NewAsync(TOPIC_ADDTOPIC, topic)); err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed to send add topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error())
					}
					if mailer := node.GetMailer(); mailer != nil {
						err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send add topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error()))
						if err != nil {
							if errorLogger := node.GetErrorLogger(); errorLogger != nil {
								errorLogger.Log(Error.New("Failed sending mail", err).Error())
							}
						}
					}
				}
			}()
		}
		return nil
	}
	return Error.New("Systemge is nil", nil)
}

// adds an async topic and its handler to systemge
// propagates the new topic of interest to all incoming connections
func (node *Node) AddAsyncTopic(topic string, handler AsyncMessageHandler) error {
	if systemge := node.systemge; systemge != nil {
		systemge.asyncMessageHandlerMutex.Lock()
		if systemge.asyncMessageHandlers[topic] != nil {
			systemge.asyncMessageHandlerMutex.Unlock()
			return Error.New("Async topic already exists", nil)
		}
		systemge.asyncMessageHandlers[topic] = handler
		systemge.asyncMessageHandlerMutex.Unlock()
		systemge.sentAddTopic.Add(1)
		systemge.incomingConnectionMutex.RLock()
		defer systemge.incomingConnectionMutex.RUnlock()
		for _, incomingConnection := range systemge.incomingConnections {
			go func() {
				if err := systemge.messageIncomingConnection(incomingConnection, Message.NewAsync(TOPIC_ADDTOPIC, topic)); err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed to send add topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error())
					}
					if mailer := node.GetMailer(); mailer != nil {
						err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send add topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error()))
						if err != nil {
							if errorLogger := node.GetErrorLogger(); errorLogger != nil {
								errorLogger.Log(Error.New("Failed sending mail", err).Error())
							}
						}
					}
				}
			}()
		}
		return nil
	}
	return Error.New("Systemge is nil", nil)
}

// removes a sync topic and its handler from systemge
// propagates the removal of the topic of interest to all incoming connections
func (node *Node) RemoveSyncTopic(topic string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.syncMessageHandlerMutex.Lock()
		if systemge.syncMessageHandlers[topic] == nil {
			systemge.syncMessageHandlerMutex.Unlock()
			return Error.New("Sync topic does not exist", nil)
		}
		delete(systemge.syncMessageHandlers, topic)
		systemge.syncMessageHandlerMutex.Unlock()
		systemge.sentRemoveTopic.Add(1)
		systemge.incomingConnectionMutex.RLock()
		defer systemge.incomingConnectionMutex.RUnlock()
		for _, incomingConnection := range systemge.incomingConnections {
			go func() {
				if err := systemge.messageIncomingConnection(incomingConnection, Message.NewAsync(TOPIC_REMOVETOPIC, topic)); err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed to send remove topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error())
					}
					if mailer := node.GetMailer(); mailer != nil {
						err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send remove topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error()))
						if err != nil {
							if errorLogger := node.GetErrorLogger(); errorLogger != nil {
								errorLogger.Log(Error.New("Failed sending mail", err).Error())
							}
						}
					}
				}
			}()
		}
		return nil
	}
	return Error.New("Systemge is nil", nil)
}

// removes an async topic and its handler from systemge
// propagates the removal of the topic of interest to all incoming connections
func (node *Node) RemoveAsyncTopic(topic string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.asyncMessageHandlerMutex.Lock()
		if systemge.asyncMessageHandlers[topic] == nil {
			systemge.asyncMessageHandlerMutex.Unlock()
			return Error.New("Async topic does not exist", nil)
		}
		delete(systemge.asyncMessageHandlers, topic)
		systemge.asyncMessageHandlerMutex.Unlock()
		systemge.sentRemoveTopic.Add(1)
		systemge.incomingConnectionMutex.RLock()
		defer systemge.incomingConnectionMutex.RUnlock()
		for _, incomingConnection := range systemge.incomingConnections {
			go func() {
				if err := systemge.messageIncomingConnection(incomingConnection, Message.NewAsync(TOPIC_REMOVETOPIC, topic)); err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed to send remove topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error())
					}
					if mailer := node.GetMailer(); mailer != nil {
						err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send remove topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error()))
						if err != nil {
							if errorLogger := node.GetErrorLogger(); errorLogger != nil {
								errorLogger.Log(Error.New("Failed sending mail", err).Error())
							}
						}
					}
				}
			}()
		}
		return nil
	}
	return Error.New("Systemge is nil", nil)
}
