package Tools

import (
	"github.com/neutralusername/Systemge/Config"
)

type ReceptionHandlerFactory[C any] func() ReceptionHandler[C]

type ReceptionHandler[C any] func([]byte, C) error

type ByteHandler[C any] func([]byte, C) error
type ObjectDeserializer[O any, C any] func([]byte, C) (O, error)
type ObjectHandler[O any, C any] func(O, C) error

func NewReceptionHandlerFactory[C any](
	receptionHandler ReceptionHandler[C],
) ReceptionHandlerFactory[C] {
	return func() ReceptionHandler[C] {
		return receptionHandler
	}
}

func NewReceptionHandler[O any, C any](
	byteHandler ByteHandler[C],
	deserializer ObjectDeserializer[O, C],
	objectHandler ObjectHandler[O, C],
) ReceptionHandler[C] {
	return func(bytes []byte, caller C) error {

		err := byteHandler(bytes, caller)
		if err != nil {
			return err
		}

		object, err := deserializer(bytes, caller)
		if err != nil {
			return err
		}

		return objectHandler(object, caller)
	}
}

func AssembleNewReceptionManagerFactory[O any, C any](
	byteRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageRateLimiterConfig *Config.TokenBucketRateLimiter,

	messageValidator ObjectHandler[O, C],
	deserializer ObjectDeserializer[O, C],

	priorityQueue *PriorityTokenQueue[O],
	obtainEnqueueConfigs ObtainEnqueueConfigs[O, C],
) ReceptionHandlerFactory[C] {

	/* byteHandlers := []ByteHandler[C]{}
	if byteRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, NewByteRateLimitByteHandler[C](NewTokenBucketRateLimiter(byteRateLimiterConfig)))
	}
	if messageRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, NewMessageRateLimitByteHandler[C](NewTokenBucketRateLimiter(messageRateLimiterConfig)))
	} */

	objectHandlers := []ObjectHandler[O, C]{}
	if messageValidator != nil {
		objectHandlers = append(objectHandlers, messageValidator)
	}
	if priorityQueue != nil && obtainEnqueueConfigs != nil {
		objectHandlers = append(objectHandlers, NewQueueObjectHandler(priorityQueue, obtainEnqueueConfigs))
	}

	return NewReceptionHandlerFactory(
		NewReceptionHandler(
			NewChainByteHandler(
				byteHandlers...,
			),
			deserializer,
			NewChainObjectHandler(
				objectHandlers...,
			),
		),
	)
}

// executes all handlers in order, return error if any handler returns an error
func NewChainObjectHandler[O any, C any](handlers ...ObjectHandler[O, C]) ObjectHandler[O, C] {
	return func(object O, caller C) error {
		for _, handler := range handlers {
			if err := handler(object, caller); err != nil {
				return err
			}
		}
		return nil
	}
}

type ObtainEnqueueConfigs[O any, C any] func(O, C) (token string, priority uint32, timeout uint32)

func NewQueueObjectHandler[O any, C any](
	priorityTokenQueue *PriorityTokenQueue[O],
	obtainEnqueueConfigs ObtainEnqueueConfigs[O, C],
) ObjectHandler[O, C] {
	return func(object O, caller C) error {
		token, priority, timeoutMs := obtainEnqueueConfigs(object, caller)
		return priorityTokenQueue.Push(token, object, priority, timeoutMs)
	}
}

type ObjectValidator[O any, C any] func(O, C) error

func NewValidationObjectHandler[O any, C any](validator ObjectValidator[O, C]) ObjectHandler[O, C] {
	return func(object O, caller C) error {
		return validator(object, caller)
	}
}

type ObtainTopic[O any, C any] func(O, C) string
type ResultHandler[O any, R any, C any] func(O, R, C) error

// resultHandler requires check for nil if applicable
func NewTopicObjectHandler[O any, R any, C any](
	topicManager *TopicManager[O, R],
	obtainTopic func(O) string,
	resultHandler ResultHandler[O, R, C],
) ObjectHandler[O, C] {
	return func(object O, caller C) error {
		if topicManager != nil {
			result, err := topicManager.Handle(obtainTopic(object), object)
			if err != nil {
				return err
			}
			return resultHandler(object, result, caller)
		}
		return nil
	}
}

// executes all handlers in order, return error if any handler returns an error
func NewChainByteHandler[C any](handlers ...ByteHandler[C]) ByteHandler[C] {
	return func(bytes []byte, caller C) error {
		for _, handler := range handlers {
			if err := handler(bytes, caller); err != nil {
				return err
			}
		}
		return nil
	}
}

func NewByteRateLimitByteHandler[C any](tokenBucketRateLimiter *TokenBucketRateLimiter) ByteHandler[C] {
	return func(bytes []byte, caller C) error {
		tokenBucketRateLimiter.Consume(uint64(len(bytes)))
		return nil
	}
}

func NewMessageRateLimitByteHandler[C any](tokenBucketRateLimiter *TokenBucketRateLimiter) ByteHandler[C] {
	return func(bytes []byte, caller C) error {
		tokenBucketRateLimiter.Consume(1)
		return nil
	}
}
