package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func NewWebsocketMessageReceptionHandlerFactory(
	byteRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageRateLimiterConfig *Config.TokenBucketRateLimiter,

	messageValidatorConfig *Config.MessageValidator,

	requestResponseManager *Tools.RequestResponseManager[*Message.Message],
	obtainResponseToken Tools.ObtainResponseToken[*Message.Message, *structName123],

	priorityQueue *Tools.PriorityTokenQueue[*Message.Message],
	obtainEnqueueConfigs Tools.ObtainEnqueueConfigs[*Message.Message, *structName123],
) Tools.ReceptionHandlerFactory[*structName123] {
	return NewWebsocketReceptionHandlerFactory[*Message.Message](
		byteRateLimiterConfig,
		messageRateLimiterConfig,

		func(bytes []byte, structName123 *structName123) (*Message.Message, error) {
			return Message.Deserialize(bytes)
		},

		func(message *Message.Message, structName123 *structName123) error {
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
		},

		requestResponseManager,
		obtainResponseToken,

		priorityQueue,
		obtainEnqueueConfigs,
	)
}

func NewWebsocketReceptionHandlerFactory[T any](
	byteRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageRateLimiterConfig *Config.TokenBucketRateLimiter,

	deserializer Tools.ObjectDeserializer[T, *structName123],

	objectValidator Tools.ObjectValidator[T, *structName123],

	requestResponseManager *Tools.RequestResponseManager[T],
	obtainResponseToken Tools.ObtainResponseToken[T, *structName123],

	priorityQueue *Tools.PriorityTokenQueue[T],
	obtainEnqueueConfigs Tools.ObtainEnqueueConfigs[T, *structName123],

	//topicManager *Tools.TopicManager,
) Tools.ReceptionHandlerFactory[*structName123] {

	byteHandlers := []Tools.ByteHandler[*structName123]{}
	if byteRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, Tools.NewByteRateLimitByteHandler[*structName123](Tools.NewTokenBucketRateLimiter(byteRateLimiterConfig)))
	}
	if messageRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, Tools.NewMessageRateLimitByteHandler[*structName123](Tools.NewTokenBucketRateLimiter(messageRateLimiterConfig)))
	}

	objectHandlers := []Tools.ObjectHandler[T, *structName123]{}
	if objectValidator != nil {
		objectHandlers = append(objectHandlers, Tools.NewValidationObjectHandler(objectValidator))
	}
	if requestResponseManager != nil && obtainResponseToken != nil {
		objectHandlers = append(objectHandlers, Tools.NewResponseObjectHandler(requestResponseManager, obtainResponseToken))
	}
	if priorityQueue != nil && obtainEnqueueConfigs != nil {
		objectHandlers = append(objectHandlers, Tools.NewQueueObjectHandler(priorityQueue, obtainEnqueueConfigs))
	}

	return Tools.NewReceptionHandlerFactory[*structName123](
		nil,
		nil,
		Tools.NewOnReception(
			Tools.NewChainByteHandler(byteHandlers...),
			deserializer,
			Tools.NewChainObjecthandler(objectHandlers...),
		),
	)
}
