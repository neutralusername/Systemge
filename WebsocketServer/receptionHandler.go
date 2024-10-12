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
	obtainResponseToken Tools.ObtainResponseToken[*Message.Message, *websocketServerReceptionHandlerCaller],

	priorityQueue *Tools.PriorityTokenQueue[*Message.Message],
	obtainEnqueueConfigs Tools.ObtainEnqueueConfigs[*Message.Message, *websocketServerReceptionHandlerCaller],
) Tools.ReceptionHandlerFactory[*websocketServerReceptionHandlerCaller] {
	return NewWebsocketReceptionHandlerFactory[*Message.Message](
		byteRateLimiterConfig,
		messageRateLimiterConfig,

		func(bytes []byte, receptionHandlerCaller *websocketServerReceptionHandlerCaller) (*Message.Message, error) {
			return Message.Deserialize(bytes)
		},

		func(message *Message.Message, websocketServerReceptionHandlerCaller *websocketServerReceptionHandlerCaller) error {
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

func NewWebsocketReceptionHandlerFactory[O any](
	byteRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageRateLimiterConfig *Config.TokenBucketRateLimiter,

	deserializer Tools.ObjectDeserializer[O, *websocketServerReceptionHandlerCaller],

	objectValidator Tools.ObjectValidator[O, *websocketServerReceptionHandlerCaller],

	requestResponseManager *Tools.RequestResponseManager[O],
	obtainResponseToken Tools.ObtainResponseToken[O, *websocketServerReceptionHandlerCaller],

	priorityQueue *Tools.PriorityTokenQueue[O],
	obtainEnqueueConfigs Tools.ObtainEnqueueConfigs[O, *websocketServerReceptionHandlerCaller],

	//topicManager *Tools.TopicManager,
) Tools.ReceptionHandlerFactory[*websocketServerReceptionHandlerCaller] {

	byteHandlers := []Tools.ByteHandler[*websocketServerReceptionHandlerCaller]{}
	if byteRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, Tools.NewByteRateLimitByteHandler[*websocketServerReceptionHandlerCaller](Tools.NewTokenBucketRateLimiter(byteRateLimiterConfig)))
	}
	if messageRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, Tools.NewMessageRateLimitByteHandler[*websocketServerReceptionHandlerCaller](Tools.NewTokenBucketRateLimiter(messageRateLimiterConfig)))
	}

	objectHandlers := []Tools.ObjectHandler[O, *websocketServerReceptionHandlerCaller]{}
	if objectValidator != nil {
		objectHandlers = append(objectHandlers, Tools.NewValidationObjectHandler(objectValidator))
	}
	if requestResponseManager != nil && obtainResponseToken != nil {
		objectHandlers = append(objectHandlers, Tools.NewResponseObjectHandler(requestResponseManager, obtainResponseToken))
	}
	if priorityQueue != nil && obtainEnqueueConfigs != nil {
		objectHandlers = append(objectHandlers, Tools.NewQueueObjectHandler(priorityQueue, obtainEnqueueConfigs))
	}

	return Tools.NewReceptionHandlerFactory[*websocketServerReceptionHandlerCaller](
		nil,
		nil,
		Tools.NewOnReception(
			Tools.NewChainByteHandler(byteHandlers...),
			deserializer,
			Tools.NewChainObjecthandler(objectHandlers...),
		),
	)
}
