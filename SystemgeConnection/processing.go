package SystemgeConnection

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (receiver *SystemgeReceiver) processingLoopSequentially() {
	if receiver.infoLogger != nil {
		receiver.infoLogger.Log("Starting processing messages sequentially")
	}
	for process := range receiver.processingChannel {
		if process == nil {
			if receiver.infoLogger != nil {
				receiver.infoLogger.Log("Stopping processing messages sequentially")
			}
			return
		}
		if len(receiver.processingChannel) >= cap(receiver.processingChannel)-1 {
			if receiver.errorLogger != nil {
				receiver.errorLogger.Log("Processing channel capacity reached")
			}
			if receiver.mailer != nil {
				err := receiver.mailer.Send(Tools.NewMail(nil, "error", Error.New("processing channel capacity reached", nil).Error()))
				if err != nil {
					if receiver.errorLogger != nil {
						receiver.errorLogger.Log(Error.New("failed sending mail", err).Error())
					}
				}
			}
		}
		process()
	}
}

func (receiver *SystemgeReceiver) processingLoopConcurrently() {
	if receiver.infoLogger != nil {
		receiver.infoLogger.Log("Starting processing messages concurrently")
	}
	for process := range receiver.processingChannel {
		if process == nil {
			if receiver.infoLogger != nil {
				receiver.infoLogger.Log("Stopping processing messages concurrently")
			}
			return
		}
		go process()
	}
}

func (receiver *SystemgeReceiver) processMessage(clientConnection *SystemgeConnection, messageBytes []byte, messageId uint32) error {
	if infoLogger := receiver.infoLogger; infoLogger != nil {
		infoLogger.Log("Processing message #" + Helpers.Uint32ToString(messageId))
	}
	defer receiver.waitGroup.Done()
	if err := receiver.checkRateLimits(messageBytes); err != nil {
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
			receiver.invalidMessagesReceived.Add(1)
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
	if infoLogger := receiver.infoLogger; infoLogger != nil {
		infoLogger.Log("Processed message #" + Helpers.Uint32ToString(messageId))
	}
	return nil
}

func (receiver *SystemgeReceiver) checkRateLimits(messageBytes []byte) error {
	if receiver.rateLimiterBytes != nil && !receiver.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		receiver.byteRateLimiterExceeded.Add(1)
		return Error.New("receiver rate limiter bytes exceeded", nil)
	}
	if receiver.rateLimiterMessages != nil && !receiver.rateLimiterMessages.Consume(1) {
		receiver.messageRateLimiterExceeded.Add(1)
		return Error.New("receiver rate limiter messages exceeded", nil)
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
