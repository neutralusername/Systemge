package reader

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

// executes all handlers in order, return error if any handler returns an error
func NewChainedReadHandler[T any](handlers ...systemge.ReadHandlerWithError[T]) systemge.ReadHandlerWithError[T] {
	return func(data T, caller systemge.Connection[T]) error {
		for _, handler := range handlers {
			if err := handler(data, caller); err != nil {
				return err
			}
		}
		return nil
	}
}

type ObtainReadHandlerEnqueueConfigs[T any] func(T, systemge.Connection[T]) (token string, priority uint32, timeoutNs int64)

type ReadHandlerQueueWrapper[T any] struct {
	Object T
	Caller systemge.Connection[T]
}

func NewQueueObjectHandler[T any](
	priorityTokenQueue *tools.PriorityTokenQueue[*ReadHandlerQueueWrapper[T]],
	obtainEnqueueConfigs ObtainReadHandlerEnqueueConfigs[T],
) systemge.ReadHandlerWithError[T] {
	return func(object T, caller systemge.Connection[T]) error {
		token, priority, timeoutNs := obtainEnqueueConfigs(object, caller)
		queueWrapper := &ReadHandlerQueueWrapper[T]{
			object,
			caller,
		}
		return priorityTokenQueue.Push(token, queueWrapper, priority, timeoutNs)
	}
}

type ObjectValidator[T any] func(T, systemge.Connection[T]) error

func NewValidationObjectHandler[T any](validator ObjectValidator[T]) systemge.ReadHandlerWithError[T] {
	return func(object T, caller systemge.Connection[T]) error {
		return validator(object, caller)
	}
}

type ObtainTokensFromBytes[T any] func(T) uint64

func NewTokenBucketRateLimitHandler[T any](
	obtainTokensFromBytes ObtainTokensFromBytes[T],
	tokenBucketRateLimiterConfig *configs.TokenBucketRateLimiter,
) systemge.ReadHandlerWithError[T] {

	tokenBucketRateLimiter := tools.NewTokenBucketRateLimiter(tokenBucketRateLimiterConfig)
	return func(bytes T, caller systemge.Connection[T]) error {
		tokens := obtainTokensFromBytes(bytes)
		tokenBucketRateLimiter.Consume(tokens)
		return nil
	}
}
