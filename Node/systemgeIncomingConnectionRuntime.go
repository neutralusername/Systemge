package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

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
		systemge.incomingConnectionsMutex.Lock()
		for _, incomingConnection := range systemge.incomingConnections {
			go func() {
				if err := systemge.messageIncomingConnection(incomingConnection, Message.NewAsync(TOPIC_ADDTOPIC, topic)); err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed to send add topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error())
					}
				}
			}()
		}
		systemge.incomingConnectionsMutex.Unlock()
		return nil
	}
	return Error.New("Systemge is nil", nil)
}

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
		systemge.incomingConnectionsMutex.Lock()
		for _, incomingConnection := range systemge.incomingConnections {
			go func() {
				if err := systemge.messageIncomingConnection(incomingConnection, Message.NewAsync(TOPIC_ADDTOPIC, topic)); err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed to send add topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error())
					}
				}
			}()
		}
		systemge.incomingConnectionsMutex.Unlock()
		return nil
	}
	return Error.New("Systemge is nil", nil)
}

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
		systemge.incomingConnectionsMutex.Lock()
		for _, incomingConnection := range systemge.incomingConnections {
			go func() {
				if err := systemge.messageIncomingConnection(incomingConnection, Message.NewAsync(TOPIC_REMOVETOPIC, topic)); err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed to send remove topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error())
					}
				}
			}()
		}
		systemge.incomingConnectionsMutex.Unlock()
		return nil
	}
	return Error.New("Systemge is nil", nil)
}

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
		systemge.incomingConnectionsMutex.Lock()
		for _, incomingConnection := range systemge.incomingConnections {
			go func() {
				if err := systemge.messageIncomingConnection(incomingConnection, Message.NewAsync(TOPIC_REMOVETOPIC, topic)); err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed to send remove topic message to incoming node connection \""+incomingConnection.name+"\"", err).Error())
					}
				}
			}()
		}
		systemge.incomingConnectionsMutex.Unlock()
		return nil
	}
	return Error.New("Systemge is nil", nil)
}
