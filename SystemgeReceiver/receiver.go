package SystemgeReceiver

import (
	"sync"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeReceiver struct {
	connection     *SystemgeConnection.SystemgeConnection
	messageHandler *SystemgeMessageHandler.SystemgeMessageHandler
	logger         *Tools.Logger
	messageChannel chan func()
	running        bool
	waitGroup      sync.WaitGroup
}

func New(connection *SystemgeConnection.SystemgeConnection, messageHandler *SystemgeMessageHandler.SystemgeMessageHandler, logger *Tools.Logger) *SystemgeReceiver {
	receiver := &SystemgeReceiver{
		connection:     connection,
		messageHandler: messageHandler,
		messageChannel: make(chan func()),
		logger:         logger,
	}
	return receiver
}

func (receiver *SystemgeReceiver) Start() {
	receiver.running = true
	go receiver.loop()
}

func (receiver *SystemgeReceiver) Stop() {
	receiver.running = false
}

func (receiver *SystemgeReceiver) loop() {
	for receiver.running {
		messageBytes, err := receiver.connection.ReceiveMessage()
		if err != nil {
			if receiver.logger != nil {
				receiver.logger.Log(Error.New("Error receiving message", err).Error())
			}
			continue
		}
		receiver.waitGroup.Add(1)
		receiver.messageChannel <- func() {
			receiver.processMessage(receiver.connection, messageBytes)
		}
	}
}

func (receiver *SystemgeReceiver) processMessage(clientConnection *SystemgeConnection.SystemgeConnection, messageBytes []byte) {
	defer receiver.waitGroup.Done()
	if err := receiver.checkRateLimits(clientConnection, messageBytes); err != nil {
		if warningLogger := receiver.warningLogger; warningLogger != nil {
			warningLogger.Log(Error.New("Rejected message from client connection \""+clientConnection.name+"\"", err).Error())
		}
		return
	}
	message, err := Message.Deserialize(messageBytes, clientConnection.name)
	if err != nil {
		receiver.invalidMessagesReceived.Add(1)
		if warningLogger := receiver.warningLogger; warningLogger != nil {
			warningLogger.Log(Error.New("Failed to deserialize message \""+string(messageBytes)+"\" from client connection \""+clientConnection.name+"\"", err).Error())
		}
		return
	}
	if err := receiver.validateMessage(message); err != nil {
		receiver.invalidMessagesReceived.Add(1)
		if warningLogger := receiver.warningLogger; warningLogger != nil {
			warningLogger.Log(Error.New("Failed to validate message \""+string(messageBytes)+"\" from client connection \""+clientConnection.name+"\"", err).Error())
		}
		return
	}
	if message.GetSyncTokenToken() == "" {
		receiver.asyncMessageBytesReceived.Add(uint64(len(messageBytes)))
		receiver.asyncMessagesReceived.Add(1)
		err := receiver.handleAsyncMessage(message)
		if err != nil {
			receiver.invalidMessagesReceived.Add(1)
			if warningLogger := receiver.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to handle async messag with topic \""+message.GetTopic()+"\" from client connection \""+clientConnection.name+"\"", err).Error())
			}
		} else {
			if infoLogger := receiver.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Handled async message with topic \""+message.GetTopic()+"\" from client connection \""+clientConnection.name+"\"", nil).Error())
			}
		}
	} else {
		receiver.syncRequestBytesReceived.Add(uint64(len(messageBytes)))
		receiver.syncRequestsReceived.Add(1)
		responsePayload, err := receiver.handleSyncRequest(message)
		if err != nil {
			receiver.invalidMessagesReceived.Add(1)
			if warningLogger := receiver.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to handle sync request with topic \""+message.GetTopic()+"\" from client connection \""+clientConnection.name+"\"", err).Error())
			}
			if err := receiver.messageClientConnection(clientConnection, message.NewFailureResponse(responsePayload)); err != nil {
				if warningLogger := receiver.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send failure response to client connection \""+clientConnection.name+"\"", err).Error())
				}
			} else {
				receiver.syncFailureResponsesSent.Add(1)
			}
		} else {
			if infoLogger := receiver.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Handled sync request with topic \""+message.GetTopic()+"\" from client connection \""+clientConnection.name+"\" with sync token \""+message.GetSyncTokenToken()+"\"", nil).Error())
			}
			if err := receiver.messageClientConnection(clientConnection, message.NewSuccessResponse(responsePayload)); err != nil {
				if warningLogger := receiver.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send success response to clientclient connection \""+clientConnection.name+"\"", err).Error())
				}
			} else {
				receiver.syncSuccessResponsesSent.Add(1)
			}
		}
	}
}

func (systemge *SystemgeReceiver) checkRateLimits(clientConnection *SystemgeConnection.SystemgeConnection, messageBytes []byte) error {
	if clientConnection.rateLimiterBytes != nil && !clientConnection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		systemge.byteRateLimiterExceeded.Add(1)
		return Error.New("client connection rate limiter bytes exceeded", nil)
	}
	if clientConnection.rateLimiterMsgs != nil && !clientConnection.rateLimiterMsgs.Consume(1) {
		systemge.messageRateLimiterExceeded.Add(1)
		return Error.New("client connection rate limiter messages exceeded", nil)
	}
	return nil
}
