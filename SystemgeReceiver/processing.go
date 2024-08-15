package SystemgeReceiver

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (receiver *SystemgeReceiver) processingLoopSequentially() {
	for process := range receiver.messageChannel {
		if process == nil {
			return
		}
		process()
	}
}

func (receiver *SystemgeReceiver) processingLoopConcurrently() {
	for process := range receiver.messageChannel {
		if process == nil {
			return
		}
		go process()
	}
}

func (receiver *SystemgeReceiver) receiveLoop() {
	for receiver.running {
		messageBytes, err := receiver.connection.ReceiveMessage()
		if err != nil {
			if receiver.errorLogger != nil {
				receiver.errorLogger.Log(Error.New("failed to receive message", err).Error())
			}
			continue
		}
		receiver.waitGroup.Add(1)
		receiver.messageChannel <- func() {
			err := receiver.processMessage(receiver.connection, messageBytes)
			if err != nil {
				if receiver.warningLogger != nil {
					receiver.warningLogger.Log(Error.New("failed to process message", err).Error())
				}
			}
		}
	}
}

func (receiver *SystemgeReceiver) processMessage(clientConnection *SystemgeConnection.SystemgeConnection, messageBytes []byte) error {
	defer receiver.waitGroup.Done()
	if err := receiver.checkRateLimits(clientConnection, messageBytes); err != nil {
		return Error.New("rejected message due to rate limits", err)
	}
	message, err := Message.Deserialize(messageBytes, clientConnection.GetName())
	if err != nil {
		receiver.invalidMessagesReceived.Add(1)
		return Error.New("failed to deserialize message", err)
	}
	if err := receiver.validateMessage(message); err != nil {
		receiver.invalidMessagesReceived.Add(1)
		return Error.New("failed to validate message", err)
	}
	if message.GetSyncTokenToken() == "" {
		receiver.asyncMessagesReceived.Add(1)
		err := receiver.messageHandler.HandleAsyncMessage(message)
		if err != nil {
			return Error.New("failed to handle async message", err)
		}
	} else if message.IsResponse() {
		if err := receiver.connection.AddSyncResponse(message); err != nil {
			receiver.invalidMessagesReceived.Add(1)
			return Error.New("failed to add sync response message", err)
		} else {
			receiver.syncResponsesReceived.Add(1)
		}
	} else {
		receiver.syncRequestsReceived.Add(1)
		if responsePayload, err := receiver.messageHandler.HandleSyncRequest(message); err != nil {
			if err := receiver.connection.SendMessage(message.NewFailureResponse(err.Error()).Serialize()); err != nil {
				return Error.New("failed to send failure response", err)
			}
		} else {
			if err := receiver.connection.SendMessage(message.NewSuccessResponse(responsePayload).Serialize()); err != nil {
				return Error.New("failed to send success response", err)
			}
		}
	}
	return nil
}

func (receiver *SystemgeReceiver) checkRateLimits(clientConnection *SystemgeConnection.SystemgeConnection, messageBytes []byte) error {
	if receiver.rateLimiterBytes != nil && !receiver.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		receiver.byteRateLimiterExceeded.Add(1)
		return Error.New("client connection rate limiter bytes exceeded", nil)
	}
	if receiver.rateLimiterMessages != nil && !receiver.rateLimiterMessages.Consume(1) {
		receiver.messageRateLimiterExceeded.Add(1)
		return Error.New("client connection rate limiter messages exceeded", nil)
	}
	return nil
}

func (receiver *SystemgeReceiver) validateMessage(message *Message.Message) error {
	if maxSyncTokenSize := receiver.config.MaxSyncTokenSize; maxSyncTokenSize > 0 && len(message.GetSyncTokenToken()) > maxSyncTokenSize {
		return Error.New("Message sync token exceeds maximum size", nil)
	}
	if len(message.GetTopic()) == 0 {
		return Error.New("Message missing topic", nil)
	}
	if maxTopicSize := receiver.config.MaxTopicSize; maxTopicSize > 0 && len(message.GetTopic()) > maxTopicSize {
		return Error.New("Message topic exceeds maximum size", nil)
	}
	if maxPayloadSize := receiver.config.MaxPayloadSize; maxPayloadSize > 0 && len(message.GetPayload()) > maxPayloadSize {
		return Error.New("Message payload exceeds maximum size", nil)
	}
	return nil
}
