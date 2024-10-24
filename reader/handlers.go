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

func NewByteRateLimitHandler(
	tokenBucketRateLimiterConfig *configs.TokenBucketRateLimiter,
) systemge.ReadHandlerWithError[[]byte] {

	tokenBucketRateLimiter := tools.NewTokenBucketRateLimiter(tokenBucketRateLimiterConfig)
	return func(data []byte, connection systemge.Connection[[]byte]) error {
		tokenBucketRateLimiter.Consume(uint64(len(data)))
		return nil
	}
}

func NewMessageRateLimitHandler[T any](
	tokenBucketRateLimiterConfig *configs.TokenBucketRateLimiter,
) systemge.ReadHandlerWithError[T] {

	tokenBucketRateLimiter := tools.NewTokenBucketRateLimiter(tokenBucketRateLimiterConfig)
	return func(data T, connection systemge.Connection[T]) error {
		tokenBucketRateLimiter.Consume(1)
		return nil
	}
}

func NewCustomRateLimitHandler[T any](
	tokenBucketRateLimiterConfig *configs.TokenBucketRateLimiter,
	consumeFunc func(T, systemge.Connection[T]) uint64,
) systemge.ReadHandlerWithError[T] {

	tokenBucketRateLimiter := tools.NewTokenBucketRateLimiter(tokenBucketRateLimiterConfig)
	return func(data T, connection systemge.Connection[T]) error {
		tokenBucketRateLimiter.Consume(consumeFunc(data, connection))
		return nil
	}
}

func NewResponseHandler[T any](
	getResponse func(T, systemge.Connection[T]) (T, error),
	writeTimeoutNs int64,
) systemge.ReadHandlerWithError[T] {
	return func(incomingData T, connection systemge.Connection[T]) error {
		response, err := getResponse(incomingData, connection)
		if err != nil {
			return err
		}
		return connection.Write(response, writeTimeoutNs)
	}
}

type ObjectValidator[T any] func(T, systemge.Connection[T]) error

func NewValidationObjectHandler[T any](
	validator ObjectValidator[T],
) systemge.ReadHandlerWithError[T] {
	return func(object T, connection systemge.Connection[T]) error {
		return validator(object, connection)
	}
}

type ObtainEnqueueConfigs[T any] func(T, systemge.Connection[T]) (token string, priority uint32, timeoutNs int64)

type queueWrapper[T any] struct {
	Object     T
	Connection systemge.Connection[T]
}

func NewQueueHandler[T any](
	priorityTokenQueue *tools.PriorityTokenQueue[*queueWrapper[T]],
	obtainEnqueueConfigs ObtainEnqueueConfigs[T],
) systemge.ReadHandlerWithError[T] {

	return func(object T, connection systemge.Connection[T]) error {
		token, priority, timeoutNs := obtainEnqueueConfigs(object, connection)
		queueWrapper := &queueWrapper[T]{
			object,
			connection,
		}
		return priorityTokenQueue.Push(token, queueWrapper, priority, timeoutNs)
	}
}

func NewDequeueRoutine[T any](
	priorityTokenQueue *tools.PriorityTokenQueue[*queueWrapper[T]],
	readHandler systemge.ReadHandlerWithError[T],
	dequeueRoutineConfig *configs.Routine,
) (*tools.Routine, error) {
	routine, err := tools.NewRoutine(
		func(stopChannel <-chan struct{}) {
			for {
				select {
				case <-stopChannel:
					return
				case queueWrapper, ok := <-priorityTokenQueue.PopChannel():
					if !ok {
						return
					}
					err := readHandler(queueWrapper.Object, queueWrapper.Connection)
					if err != nil {

					}
				}
			}
		},
		dequeueRoutineConfig,
	)
	if err != nil {
		return nil, err
	}
	return routine, nil
}

/* type messageHandlerWrapper[T any] struct {
	message    tools.IMessage
	connection systemge.Connection[T]
}

func NewMessageHandler[T any](
	asyncMessageHandlers systemge.AsyncMessageHandlers[T],
	unknownAsyncMessageHandler systemge.AsyncMessageHandler[T],
	syncMessageHandlers systemge.SyncMessageHandlers[T],
	unknownSyncMessageHandler systemge.SyncMessageHandler[T],
	asyncTopicManagerConfig *configs.TopicManager,
	syncTopicManagerConfig *configs.TopicManager,
) systemge.ReadHandlerWithError[T] {
	asyncTopicHandlers := tools.TopicHandlers[messageHandlerWrapper[T], T]{}

	for topic, handler := range asyncMessageHandlers {
		asyncTopicHandlers[topic] = func(mhw messageHandlerWrapper[T]) (T, error) {
			handler(mhw.connection, mhw.message)
			var nilValue T
			return nilValue, nil
		}
	}

	topicManagerAsync := tools.NewTopicManager[T, *T](
		asyncTopicManagerConfig,
		tools.TopicHandlers[T, *T]{},
		nil,
	)
	topicManagerSync := tools.NewTopicManager[T, T](
		syncTopicManagerConfig,
		tools.TopicHandlers[T, T]{},
		nil,
	)

	return func(message T, connection systemge.Connection[T]) error {

	}
}
*/
