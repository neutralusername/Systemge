package Tools

import (
	"github.com/neutralusername/Systemge/Config"
)

type ReceptionHandlerFactory[S any] func() ReceptionHandler[S]
type ReceptionHandler[S any] func([]byte, S) error

type ByteHandler[S any] func([]byte, S) error
type ObjectDeserializer[T any, S any] func([]byte, S) (T, error)
type ObjectHandler[T any, S any] func(T, S) error

func NewReceptionHandler[T any, S any](
	byteHandler ByteHandler[S],
	deserializer ObjectDeserializer[T, S],
	objectHandler ObjectHandler[T, S],
) ReceptionHandler[S] {
	return func(bytes []byte, structName123 S) error {

		err := byteHandler(bytes, structName123)
		if err != nil {
			return err
		}

		object, err := deserializer(bytes, structName123)
		if err != nil {
			return err
		}

		return objectHandler(object, structName123)
	}
}

type ObtainEnqueueConfigs[T any, S any] func(T, S) (token string, priority uint32, timeout uint32)
type ObtainResponseToken[T any, S any] func(T, S) string
type ObjectValidator[T any, S any] func(T, S) error
type ObtainTopic[T any, S any] func(T, S) string
type ResultHandler[T any, R any, S any] func(T, R, S) error

// executes all handlers in order, return error if any handler returns an error
func NewChainObjecthandler[T any, S any](handlers ...ObjectHandler[T, S]) ObjectHandler[T, S] {
	return func(object T, structName123 S) error {
		for _, handler := range handlers {
			if err := handler(object, structName123); err != nil {
				return err
			}
		}
		return nil
	}
}

func NewQueueObjectHandler[T any, S any](
	priorityTokenQueue *PriorityTokenQueue[T],
	obtainEnqueueConfigs ObtainEnqueueConfigs[T, S],
) ObjectHandler[T, S] {
	return func(object T, structName123 S) error {
		token, priority, timeoutMs := obtainEnqueueConfigs(object, structName123)
		return priorityTokenQueue.Push(token, object, priority, timeoutMs)
	}
}

func NewResponseObjectHandler[T any, S any](
	requestResponseManager *RequestResponseManager[T],
	obtainResponseToken ObtainResponseToken[T, S],
) ObjectHandler[T, S] {
	return func(object T, structName123 S) error {
		responseToken := obtainResponseToken(object, structName123)
		if responseToken != "" {
			if requestResponseManager != nil {
				return requestResponseManager.AddResponse(responseToken, object)
			}
		}
		return nil
	}
}

// resultHandler requires check for nil if applicable
func NewTopicObjectHandler[T any, R any, S any](
	topicManager *TopicManager[T, R],
	obtainTopic func(T) string,
	resultHandler ResultHandler[T, R, S],
) ObjectHandler[T, S] {
	return func(object T, structName123 S) error {
		if topicManager != nil {
			result, err := topicManager.Handle(obtainTopic(object), object)
			if err != nil {
				return err
			}
			return resultHandler(object, result, structName123)
		}
		return nil
	}
}

func NewValidationObjectHandler[T any, S any](validator ObjectValidator[T, S]) ObjectHandler[T, S] {
	return func(object T, structName123 S) error {
		return validator(object, structName123)
	}
}

// executes all handlers in order, return error if any handler returns an error
func NewChainByteHandler[S any](handlers ...ByteHandler[S]) ByteHandler[S] {
	return func(bytes []byte, structName123 S) error {
		for _, handler := range handlers {
			if err := handler(bytes, structName123); err != nil {
				return err
			}
		}
		return nil
	}
}

func NewByteRateLimitByteHandler[S any](tokenBucketRateLimiter *TokenBucketRateLimiter) ByteHandler[S] {
	return func(bytes []byte, structName123 S) error {
		tokenBucketRateLimiter.Consume(uint64(len(bytes)))
		return nil
	}
}

func NewMessageRateLimitByteHandler[S any](tokenBucketRateLimiter *TokenBucketRateLimiter) ByteHandler[S] {
	return func(bytes []byte, structName123 S) error {
		tokenBucketRateLimiter.Consume(1)
		return nil
	}
}

func NewDefaultReceptionHandlerFactory[S any]() ReceptionHandlerFactory[S] {
	return func() ReceptionHandler[S] {

		return func(bytes []byte, structName123 S) error {
			return nil
		}
	}
}

func NewValidationMessageReceptionHandlerFactory[T any, S any](
	byteRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageValidatorConfig *Config.MessageValidator,
	//topicManager *Tools.TopicManager,
	priorityQueue *PriorityTokenQueue[T],
	obtainEnqueueConfigs ObtainEnqueueConfigs[T, S],
	requestResponseManager *RequestResponseManager[T],
) ReceptionHandlerFactory[S] {

	byteHandlers := []ByteHandler[S]{}
	if byteRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, NewByteRateLimitByteHandler[S](NewTokenBucketRateLimiter(byteRateLimiterConfig)))
	}
	if messageRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, NewMessageRateLimitByteHandler[S](NewTokenBucketRateLimiter(messageRateLimiterConfig)))
	}

	objectDeserializer := func(messageBytes []byte, s S) (T, error) {
		/* 	return Message.Deserialize(messageBytes) */
	}

	objectHandlers := []ObjectHandler[T, S]{}
	if messageValidatorConfig != nil {
		objectValidator := func(message T, s S) error {
			/* if messageValidatorConfig.MinSyncTokenSize >= 0 && len(message.GetSyncToken()) < messageValidatorConfig.MinSyncTokenSize {
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
			} */
			return nil
		}
		objectHandlers = append(objectHandlers, NewValidationObjectHandler(objectValidator))
	}
	if requestResponseManager != nil {
		obtainResponseToken := func(message T, s S) string {
			/* if message.IsResponse() {
				return message.GetSyncToken()
			} */
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

func NewValidationReceptionHandlerFactory[T any, S any](
	byteHandler ByteHandler[S],
	deserializer ObjectDeserializer[T, S],
	objectHandler ObjectHandler[T, S],
) ReceptionHandlerFactory[S] {

	return func() ReceptionHandler[S] {

		return NewReceptionHandler[T, S](
			byteHandler,
			deserializer,
			objectHandler,
		)
	}
}
