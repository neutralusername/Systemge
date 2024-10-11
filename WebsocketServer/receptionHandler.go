package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func NewDefaultReceptionHandlerFactory() Tools.ReceptionHandlerFactory[*structName123] {
	return func() Tools.ReceptionHandler[*structName123] {

		return func(bytes []byte, structName123 *structName123) error {
			return nil
		}
	}
}

func NewValidationMessageReceptionHandlerFactory(
	byteRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageValidatorConfig *Config.MessageValidator,
	//topicManager *Tools.TopicManager,
	priorityQueue *Tools.PriorityTokenQueue[*Message.Message],
	obtainEnqueueConfigs Tools.ObtainEnqueueConfigs[*Message.Message],
	requestResponseManager *Tools.RequestResponseManager[*Message.Message],
) Tools.ReceptionHandlerFactory[*structName123] {

	byteHandlers := []Tools.ByteHandler[*Message.Message]{}
	if byteRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, Tools.NewByteRateLimitByteHandler[*Message.Message](Tools.NewTokenBucketRateLimiter(byteRateLimiterConfig)))
	}
	if messageRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, Tools.NewMessageRateLimitByteHandler[*Message.Message](Tools.NewTokenBucketRateLimiter(messageRateLimiterConfig)))
	}

	objectDeserializer := func(messageBytes []byte) (*Message.Message, error) {
		return Message.Deserialize(messageBytes)
	}

	objectHandlers := []Tools.ObjectHandler[*Message.Message]{}
	if messageValidatorConfig != nil {
		objectValidator := func(message *Message.Message) error {
			if messageValidatorConfig.MinSyncTokenSize >= 0 && len(message.GetSyncToken()) < messageValidatorConfig.MinSyncTokenSize {
				return errors.New("message contains sync token")
			}
			if messageValidatorConfig.MaxSyncTokenSize >= 0 && len(message.GetSyncToken()) > messageValidatorConfig.MaxSyncTokenSize {
				return errors.New("message contains sync token")
			}
			if messageValidatorConfig.MinTopicSize >= 0 && len(message.GetTopic()) < messageValidatorConfig.MinTopicSize {
				return errors.New("message missing topic")
			}
			if messageValidatorConfig.MaxTopicSize >= 0 && len(message.GetTopic()) > messageValidatorConfig.MaxTopicSize {
				return errors.New("message missing topic")
			}
			if messageValidatorConfig.MinPayloadSize >= 0 && len(message.GetPayload()) < messageValidatorConfig.MinPayloadSize {
				return errors.New("message payload exceeds maximum size")
			}
			if messageValidatorConfig.MaxPayloadSize >= 0 && len(message.GetPayload()) > messageValidatorConfig.MaxPayloadSize {
				return errors.New("message payload exceeds maximum size")
			}
			return nil
		}
		objectHandlers = append(objectHandlers, Tools.NewValidationObjectHandler(objectValidator))
	}
	if requestResponseManager != nil {
		obtainResponseToken := func(message *Message.Message) string {
			if message.IsResponse() {
				return message.GetSyncToken()
			}
			return ""
		}
		objectHandlers = append(objectHandlers, Tools.NewResponseObjectHandler(requestResponseManager, obtainResponseToken))
	}
	if priorityQueue != nil {
		if obtainEnqueueConfigs == nil {
			obtainEnqueueConfigs = func(message *Message.Message) (string, uint32, uint32) {
				return "", 0, 0
			}
		}
		objectHandlers = append(objectHandlers, Tools.NewQueueObjectHandler(priorityQueue, obtainEnqueueConfigs))

	}

	return NewValidationReceptionHandlerFactory[*Message.Message](
		Tools.NewChainByteHandler(byteHandlers...),
		objectDeserializer,
		Tools.NewChainObjecthandler(objectHandlers...),
	)
}

func NewValidationReceptionHandlerFactory[T any](
	byteHandler Tools.ByteHandler[T],
	deserializer Tools.ObjectDeserializer[T],
	objectHandler Tools.ObjectHandler[T],
) Tools.ReceptionHandlerFactory[*structName123] {

	return func() Tools.ReceptionHandler[*structName123] {

		return Tools.NewReceptionHandler[T, *structName123](
			byteHandler,
			deserializer,
			objectHandler,
		)
	}
}
