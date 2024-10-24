package reader

import (
	"errors"

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

type messageHandlerWrapper[T any] struct {
	message    tools.IMessage
	connection systemge.Connection[T]
}

func NewMessageTopicManager[T any](
	asyncMessageHandlers systemge.AsyncMessageHandlers[T],
	syncMessageHandlers systemge.SyncMessageHandlers[T],
	unknownAsyncMessageHandler systemge.AsyncMessageHandler[T],
	unknownSyncMessageHandler systemge.SyncMessageHandler[T],
	topicManagerConfig *configs.TopicManager,
) *tools.TopicManager[messageHandlerWrapper[T]] {

	topicHandlers := tools.TopicHandlers[messageHandlerWrapper[T]]{}
	for topic, handler := range asyncMessageHandlers {
		topicHandlers[topic] = func(mhw messageHandlerWrapper[T]) {
			handler(mhw.connection, mhw.message)
		}
	}
	for topic, handler := range syncMessageHandlers {
		topicHandlers[topic] = func(mhw messageHandlerWrapper[T]) {
			response, err := handler(mhw.connection, mhw.message)
			if err != nil {
				// do smthg w the err
				return
			}
			err = mhw.connection.Write(response, 0)
			if err != nil {
				// do smthg w the err
			}
		}
	}

	var unknownTopicHandler tools.TopicHandler[messageHandlerWrapper[T]]
	if unknownAsyncMessageHandler != nil || unknownSyncMessageHandler != nil {
		unknownTopicHandler = func(mhw messageHandlerWrapper[T]) {
			if unknownAsyncMessageHandler != nil {
				unknownAsyncMessageHandler(mhw.connection, mhw.message)
			}
			if unknownSyncMessageHandler != nil {
				response, err := unknownSyncMessageHandler(mhw.connection, mhw.message)
				if err != nil {
					// do smthg w the err
					return
				}
				err = mhw.connection.Write(response, 0)
				if err != nil {
					// do smthg w the err
				}
			}
		}
	}

	topicManager := tools.NewTopicManager[messageHandlerWrapper[T]](
		topicManagerConfig,
		topicHandlers,
		unknownTopicHandler,
	)

	return topicManager
}

func NewTopicMessageHandler[T any](
	topicManager *tools.TopicManager[messageHandlerWrapper[T]],
	retrieveMessage func(T, systemge.Connection[T]) tools.IMessage,
) systemge.ReadHandlerWithError[T] {

	return func(data T, connection systemge.Connection[T]) error {
		message := retrieveMessage(data, connection)
		if message == nil {
			return errors.New("could not retrieve message")
		}
		return topicManager.Handle(message.GetTopic(), messageHandlerWrapper[T]{
			message,
			connection,
		})
	}
}
