package ReceptionHandler

import (
	"github.com/neutralusername/Systemge/Tools"
)

type ReceptionHandler func([]byte) error

type ByteHandler[T any] func([]byte) error
type ObjectDeserializer[T any] func([]byte) (T, error)

type ObjectHandler[T any] func(T) error

type ObtainEnqueueConfigs[T any] func(T) (string, uint32, uint32)
type ObtainResponseToken[T any] func(T) string
type ObjectValidator[T any] func(T) error

func NewQueueObjectHandler[T any](
	priorityTokenQueue *Tools.PriorityTokenQueue[T],
	obtainEnqueueConfigs ObtainEnqueueConfigs[T],
) ObjectHandler[T] {
	return func(object T) error {
		token, priority, timeoutMs := obtainEnqueueConfigs(object)
		return priorityTokenQueue.Push(token, object, priority, timeoutMs)
	}
}

func NewResponseObjectHandler[T any](
	requestResponseManager *Tools.RequestResponseManager[T],
	obtainResponseToken ObtainResponseToken[T],
) ObjectHandler[T] {
	return func(object T) error {
		responseToken := obtainResponseToken(object)
		if responseToken != "" {
			if requestResponseManager != nil {
				return requestResponseManager.AddResponse(responseToken, object)
			}
		}
		return nil
	}
}

func NewValidationObjectHandler[T any](
	validator ObjectValidator[T],
) ObjectHandler[T] {
	return func(object T) error {
		return validator(object)
	}
}

// executes all handlers in order, return error if any handler returns an error
func NewChainObjecthandler[T any](
	handlers ...ObjectHandler[T],
) ObjectHandler[T] {
	return func(object T) error {
		for _, handler := range handlers {
			if err := handler(object); err != nil {
				return err
			}
		}
		return nil
	}
}

func NewChainByteHandler[T any](
	handlers ...ByteHandler[T],
) ByteHandler[T] {
	return func(bytes []byte) error {
		for _, handler := range handlers {
			if err := handler(bytes); err != nil {
				return err
			}
		}
		return nil
	}
}

func NewByteRateLimitByteHandler[T any](
	tokenBucketRateLimiter *Tools.TokenBucketRateLimiter,
) ByteHandler[T] {
	return func(bytes []byte) error {
		tokenBucketRateLimiter.Consume(uint64(len(bytes)))
		return nil
	}
}

func NewMessageRateLimitByteHandler[T any](
	tokenBucketRateLimiter *Tools.TokenBucketRateLimiter,
) ByteHandler[T] {
	return func(bytes []byte) error {
		tokenBucketRateLimiter.Consume(1)
		return nil
	}
}

func NewReceptionHandler[T any](
	byteHandler ByteHandler[T],
	deserializer ObjectDeserializer[T],
	objectHandler ObjectHandler[T],
) ReceptionHandler {
	return func(bytes []byte) error {

		err := byteHandler(bytes)
		if err != nil {
			return err
		}

		object, err := deserializer(bytes)
		if err != nil {
			return err
		}

		return objectHandler(object)
	}
}
