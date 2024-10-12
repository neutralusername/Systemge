package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func NewWebsocketMessageReceptionManagerFactory(
	byteRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageRateLimiterConfig *Config.TokenBucketRateLimiter,

	messageValidatorConfig *Config.MessageValidator,

	requestResponseManager *Tools.RequestResponseManager[*Message.Message],
	obtainResponseToken Tools.ObtainResponseToken[*Message.Message, *websocketServerReceptionManagerCaller],

	priorityQueue *Tools.PriorityTokenQueue[*Message.Message],
	obtainEnqueueConfigs Tools.ObtainEnqueueConfigs[*Message.Message, *websocketServerReceptionManagerCaller],
) Tools.ReceptionManagerFactory[*websocketServerReceptionManagerCaller] {
	return NewWebsocketReceptionManagerFactory[*Message.Message](
		byteRateLimiterConfig,
		messageRateLimiterConfig,

		func(bytes []byte, caller *websocketServerReceptionManagerCaller) (*Message.Message, error) {
			return Message.Deserialize(bytes)
		},

		func(message *Message.Message, caller *websocketServerReceptionManagerCaller) error {
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

		nil,
		nil,
	)
}

func NewWebsocketReceptionManagerFactory(
	/* byteRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageRateLimiterConfig *Config.TokenBucketRateLimiter,

	deserializer Tools.ObjectDeserializer[O, *websocketServerReceptionManagerCaller],

	objectValidator Tools.ObjectValidator[O, *websocketServerReceptionManagerCaller],

	requestResponseManager *Tools.RequestResponseManager[O],
	obtainResponseToken Tools.ObtainResponseToken[O, *websocketServerReceptionManagerCaller],

	priorityQueue *Tools.PriorityTokenQueue[O],
	obtainEnqueueConfigs Tools.ObtainEnqueueConfigs[O, *websocketServerReceptionManagerCaller], */

	onStart Tools.OnReceptionManagerStart[*websocketServerReceptionManagerCaller],
	onStop Tools.OnReceptionManagerStop[*websocketServerReceptionManagerCaller],
	onHandle Tools.OnReceptionManagerHandle[*websocketServerReceptionManagerCaller],
) Tools.ReceptionManagerFactory[*websocketServerReceptionManagerCaller] {

	/* 	byteHandlers := []Tools.ByteHandler[*websocketServerReceptionManagerCaller]{}
	   	if byteRateLimiterConfig != nil {
	   		byteHandlers = append(byteHandlers, Tools.NewByteRateLimitByteHandler[*websocketServerReceptionManagerCaller](Tools.NewTokenBucketRateLimiter(byteRateLimiterConfig)))
	   	}
	   	if messageRateLimiterConfig != nil {
	   		byteHandlers = append(byteHandlers, Tools.NewMessageRateLimitByteHandler[*websocketServerReceptionManagerCaller](Tools.NewTokenBucketRateLimiter(messageRateLimiterConfig)))
	   	} */

	/* 	objectHandlers := []Tools.ObjectHandler[O, *websocketServerReceptionManagerCaller]{}
	   	if objectValidator != nil {
	   		objectHandlers = append(objectHandlers, Tools.NewValidationObjectHandler(objectValidator))
	   	}
	   	if requestResponseManager != nil && obtainResponseToken != nil {
	   		objectHandlers = append(objectHandlers, Tools.NewResponseObjectHandler(requestResponseManager, obtainResponseToken))
	   	}
	   	if priorityQueue != nil && obtainEnqueueConfigs != nil {
	   		objectHandlers = append(objectHandlers, Tools.NewQueueObjectHandler(priorityQueue, obtainEnqueueConfigs))
	   	} */

	return Tools.NewReceptionManagerFactory[*websocketServerReceptionManagerCaller](
		onStart,
		onStop,
		onHandle,
		/* 		Tools.NewOnReceptionManagerHandle(
			Tools.NewChainByteHandler(byteHandlers...),
			deserializer,
			Tools.NewChainObjecthandler(objectHandlers...),
		), */
	)
}
