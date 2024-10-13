package Tools

import (
	"github.com/neutralusername/Systemge/Config"
)

type ReadHandlerFactory[C any] func() ReadHandler[C]

type ReadHandler[C any] func([]byte, C)

type ByteHandler[C any] func([]byte, C) error
type ObjectDeserializer[O any, C any] func([]byte, C) (O, error)
type ObjectHandler[O any, C any] func(O, C) error

type QueueWrapper[O any, C any] struct {
	object O
	caller C
}

func NewReadHandler[O any, C any](
	byteHandler ByteHandler[C],
	deserializer ObjectDeserializer[O, C],
	objectHandler ObjectHandler[O, C],
) ReadHandler[C] {
	return func(bytes []byte, caller C) {

		err := byteHandler(bytes, caller)
		if err != nil {
			return
		}

		object, err := deserializer(bytes, caller)
		if err != nil {
			return
		}

		objectHandler(object, caller)
	}
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
	priorityTokenQueue *PriorityTokenQueue[*QueueWrapper[O, C]],
	obtainEnqueueConfigs ObtainEnqueueConfigs[O, C],
) ObjectHandler[O, C] {
	return func(object O, caller C) error {
		token, priority, timeoutMs := obtainEnqueueConfigs(object, caller)
		queueWrapper := &QueueWrapper[O, C]{
			object,
			caller,
		}
		return priorityTokenQueue.Push(token, queueWrapper, priority, timeoutMs)
	}
}

type ObjectValidator[O any, C any] func(O, C) error

func NewValidationObjectHandler[O any, C any](validator ObjectValidator[O, C]) ObjectHandler[O, C] {
	return func(object O, caller C) error {
		return validator(object, caller)
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

func NewTokenBucketRateLimitHandler[C any](tokenBucketRateLimiterConfig *Config.TokenBucketRateLimiter) ByteHandler[C] {
	tokenBucketRateLimiter := NewTokenBucketRateLimiter(tokenBucketRateLimiterConfig)
	return func(bytes []byte, caller C) error {
		tokenBucketRateLimiter.Consume(uint64(len(bytes)))
		return nil
	}
}
