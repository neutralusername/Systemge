package SystemgeServer

import (
	"sync"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

// handles client messages from a client connection one at a time until the receive operation fails
// either due to connection loss or closure of the listener due to systemge stop
func (systemge *SystemgeServer) handleClientConnectionMessages(clientConnection *clientConnection) {
	if infoLogger := systemge.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Starting message handler for client connection \""+clientConnection.name+"\"", nil).Error())
	}
	wg := sync.WaitGroup{}
	for {
		messageBytes, err := clientConnection.receiveMessage(systemge.config.TcpBufferBytes, systemge.config.ClientMessageByteLimit)
		if err != nil {
			if warningLogger := systemge.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from client ne connection \""+clientConnection.name+"\" likely due to connection loss", err).Error())
			}
			wg.Wait()
			clientConnection.netConn.Close()
			if clientConnection.rateLimiterBytes != nil {
				clientConnection.rateLimiterBytes.Stop()
			}
			if clientConnection.rateLimiterMsgs != nil {
				clientConnection.rateLimiterMsgs.Stop()
			}
			close(clientConnection.stopChannel)
			systemge.clientConnectionMutex.Lock()
			delete(systemge.clientConnections, clientConnection.name)
			systemge.clientConnectionMutex.Unlock()
			return
		}
		wg.Add(1)
		if systemge.config.ProcessAllMessagesSequentially {
			systemge.messageHandlerChannel <- func() {
				systemge.processClientMessage(clientConnection, messageBytes, &wg)
			}
		} else if !systemge.config.ProcessMessagesOfEachConnectionSequentially {
			go systemge.processClientMessage(clientConnection, messageBytes, &wg)
		} else {
			systemge.processClientMessage(clientConnection, messageBytes, &wg)
		}
	}
}

func (systemge *SystemgeServer) processClientMessage(clientConnection *clientConnection, messageBytes []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	systemge.bytesReceived.Add(uint64(len(messageBytes)))
	if err := systemge.checkRateLimits(clientConnection, messageBytes); err != nil {
		if warningLogger := systemge.warningLogger; warningLogger != nil {
			warningLogger.Log(Error.New("Rejected message from client connection \""+clientConnection.name+"\"", err).Error())
		}
		return
	}
	message, err := Message.Deserialize(messageBytes, clientConnection.name)
	if err != nil {
		systemge.invalidMessagesReceived.Add(1)
		if warningLogger := systemge.warningLogger; warningLogger != nil {
			warningLogger.Log(Error.New("Failed to deserialize message \""+string(messageBytes)+"\" from client connection \""+clientConnection.name+"\"", err).Error())
		}
		return
	}
	if err := systemge.validateMessage(message); err != nil {
		systemge.invalidMessagesReceived.Add(1)
		if warningLogger := systemge.warningLogger; warningLogger != nil {
			warningLogger.Log(Error.New("Failed to validate message \""+string(messageBytes)+"\" from client connection \""+clientConnection.name+"\"", err).Error())
		}
		return
	}
	if message.GetSyncTokenToken() == "" {
		systemge.asyncMessageBytesReceived.Add(uint64(len(messageBytes)))
		systemge.asyncMessagesReceived.Add(1)
		err := systemge.handleAsyncMessage(message)
		if err != nil {
			systemge.invalidMessagesReceived.Add(1)
			if warningLogger := systemge.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to handle async messag with topic \""+message.GetTopic()+"\" from client connection \""+clientConnection.name+"\"", err).Error())
			}
		} else {
			if infoLogger := systemge.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Handled async message with topic \""+message.GetTopic()+"\" from client connection \""+clientConnection.name+"\"", nil).Error())
			}
		}
	} else {
		systemge.syncRequestBytesReceived.Add(uint64(len(messageBytes)))
		systemge.syncRequestsReceived.Add(1)
		responsePayload, err := systemge.handleSyncRequest(message)
		if err != nil {
			systemge.invalidMessagesReceived.Add(1)
			if warningLogger := systemge.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to handle sync request with topic \""+message.GetTopic()+"\" from client connection \""+clientConnection.name+"\"", err).Error())
			}
			if err := systemge.messageClientConnection(clientConnection, message.NewFailureResponse(responsePayload)); err != nil {
				if warningLogger := systemge.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send failure response to client connection \""+clientConnection.name+"\"", err).Error())
				}
			} else {
				systemge.syncFailureResponsesSent.Add(1)
			}
		} else {
			if infoLogger := systemge.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Handled sync request with topic \""+message.GetTopic()+"\" from client connection \""+clientConnection.name+"\" with sync token \""+message.GetSyncTokenToken()+"\"", nil).Error())
			}
			if err := systemge.messageClientConnection(clientConnection, message.NewSuccessResponse(responsePayload)); err != nil {
				if warningLogger := systemge.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send success response to clientclient connection \""+clientConnection.name+"\"", err).Error())
				}
			} else {
				systemge.syncSuccessResponsesSent.Add(1)
			}
		}
	}
}

func (systemge *SystemgeServer) handleSyncRequest(message *Message.Message) (string, error) {
	systemge.syncMessageHandlerMutex.RLock()
	syncMessageHandler := systemge.syncMessageHandlers[message.GetTopic()]
	systemge.syncMessageHandlerMutex.RUnlock()
	if syncMessageHandler == nil {
		return "Not responsible for topic \"" + message.GetTopic() + "\"", Error.New("Received sync request with topic \""+message.GetTopic()+"\" for which no handler is registered", nil)
	}
	responsePayload, err := syncMessageHandler(message)
	if err != nil {
		return err.Error(), Error.New("Sync message handler for topic \""+message.GetTopic()+"\" returned error", err)
	}
	return responsePayload, nil
}

func (systemge *SystemgeServer) handleAsyncMessage(message *Message.Message) error {
	systemge.asyncMessageHandlerMutex.RLock()
	asyncMessageHandler := systemge.asyncMessageHandlers[message.GetTopic()]
	systemge.asyncMessageHandlerMutex.RUnlock()
	if asyncMessageHandler == nil {
		return Error.New("Received async message with topic \""+message.GetTopic()+"\" for which no handler is registered", nil)
	}
	err := asyncMessageHandler(message)
	if err != nil {
		return Error.New("Async message handler for topic \""+message.GetTopic()+"\" returned error", err)
	}
	return nil
}
