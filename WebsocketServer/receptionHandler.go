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

	byteHandlers := []Tools.ByteHandler[*websocketServerReceptionManagerCaller]{}
	if byteRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, Tools.NewByteRateLimitByteHandler[*websocketServerReceptionManagerCaller](Tools.NewTokenBucketRateLimiter(byteRateLimiterConfig)))
	}
	if messageRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, Tools.NewMessageRateLimitByteHandler[*websocketServerReceptionManagerCaller](Tools.NewTokenBucketRateLimiter(messageRateLimiterConfig)))
	}

	deserializer := func(bytes []byte, caller *websocketServerReceptionManagerCaller) (*Message.Message, error) {
		return Message.Deserialize(bytes)
	}

	objectHandlers := []Tools.ObjectHandler[*Message.Message, *websocketServerReceptionManagerCaller]{}
	objectHandlers = append(objectHandlers, Tools.NewValidationObjectHandler(func(message *Message.Message, caller *websocketServerReceptionManagerCaller) error {
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
	}))
	if requestResponseManager != nil && obtainResponseToken != nil {
		objectHandlers = append(objectHandlers, Tools.NewResponseObjectHandler(requestResponseManager, obtainResponseToken))
	}
	if priorityQueue != nil && obtainEnqueueConfigs != nil {
		objectHandlers = append(objectHandlers, Tools.NewQueueObjectHandler(priorityQueue, obtainEnqueueConfigs))
	}

	return NewWebsocketReceptionManagerFactory(
		nil,
		nil,
		Tools.NewOnReceptionManagerHandle(
			Tools.NewChainByteHandler(
				byteHandlers...,
			),
			deserializer,
			Tools.NewChainObjectHandler(
				objectHandlers...,
			),
		),
	)
}

func NewWebsocketReceptionManagerFactory(
	onStart Tools.OnReceptionManagerStart[*websocketServerReceptionManagerCaller],
	onStop Tools.OnReceptionManagerStop[*websocketServerReceptionManagerCaller],
	onHandle Tools.OnReceptionManagerHandle[*websocketServerReceptionManagerCaller],
) Tools.ReceptionManagerFactory[*websocketServerReceptionManagerCaller] {
	return Tools.NewReceptionManagerFactory[*websocketServerReceptionManagerCaller](
		onStart,
		onStop,
		onHandle,
	)
}
