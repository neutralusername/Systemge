package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Tools"
)

func NewWebsocketMessageReceptionHandlerFactory[T any](
	byteRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageRateLimiterConfig *Config.TokenBucketRateLimiter,

	objectValidator Tools.ObjectValidator[T, *structName123],
	obtainEnqueueConfigs Tools.ObtainEnqueueConfigs[T, *structName123],
	//topicManager *Tools.TopicManager,
	priorityQueue *Tools.PriorityTokenQueue[T],
	requestResponseManager *Tools.RequestResponseManager[T],
) Tools.ReceptionHandlerFactory[*structName123] {

	return Tools.NewReceptionHandlerFactory[*structName123](
		Tools.NewByteRateLimitByteHandler[*structName123](Tools.NewTokenBucketRateLimiter(byteRateLimiterConfig)),
		Tools.NewMessageRateLimitByteHandler[*structName123](Tools.NewTokenBucketRateLimiter(messageRateLimiterConfig)),
		objectValidator,
		//topicManager,
		priorityQueue,
		obtainEnqueueConfigs,
		requestResponseManager,
	)
}

/*

func NewValidationMessageReceptionHandlerFactory[T any, S any](
	byteRateLimiter *TokenBucketRateLimiter,
	messageRateLimiter *TokenBucketRateLimiter,
	objectValidator ObjectValidator[T, S],
	//topicManager *Tools.TopicManager,
	priorityQueue *PriorityTokenQueue[T],
	obtainEnqueueConfigs ObtainEnqueueConfigs[T, S],
	requestResponseManager *RequestResponseManager[T],
) ReceptionHandlerFactory[S] {

	byteHandlers := []ByteHandler[S]{}
	if byteRateLimiter != nil {
		byteHandlers = append(byteHandlers, NewByteRateLimitByteHandler[S](byteRateLimiter))
	}
	if messageRateLimiter != nil {
		byteHandlers = append(byteHandlers, NewMessageRateLimitByteHandler[S](messageRateLimiter))
	}

	objectDeserializer := func(messageBytes []byte, s S) (T, error) {
		 	return Message.Deserialize(messageBytes)
	}

	objectHandlers := []ObjectHandler[T, S]{}
	if objectValidator != nil {
		objectValidator := func(message T, s S) error {
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
		objectHandlers = append(objectHandlers, NewValidationObjectHandler(objectValidator))
	}
	if requestResponseManager != nil {
		obtainResponseToken := func(message T, s S) string {
			 if message.IsResponse() {
				return message.GetSyncToken()
			}
			return ""
		}
		objectHandlers = append(objectHandlers, NewResponseObjectHandler(requestResponseManager, obtainResponseToken))
	}
	if priorityQueue != nil {
		if obtainEnqueueConfigs == nil {
			obtainEnqueueConfigs = func(message T, s S) (string, uint32, uint32) {
				return "", 0, 0
			}
		}
		objectHandlers = append(objectHandlers, NewQueueObjectHandler(priorityQueue, obtainEnqueueConfigs))

	}

	return NewValidationReceptionHandlerFactory(
		NewChainByteHandler(byteHandlers...),
		objectDeserializer,
		NewChainObjecthandler(objectHandlers...),
	)
}

*/
