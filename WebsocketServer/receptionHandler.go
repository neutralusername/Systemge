package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Tools"
)

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

	return Tools.NewReceptionHandlerFactory[T, *structName123](
		Tools.NewChainByteHandler(byteHandlers...),
		deserializer,
		Tools.NewChainObjecthandler(objectHandlers...),
	)
}
