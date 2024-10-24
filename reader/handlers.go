package reader

import (
	"errors"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type Handler[T any] func(T, systemge.Connection[T])
type HandlerWithResult[T any] func(T, systemge.Connection[T]) (T, error)
type HandlerWithError[T any] func(T, systemge.Connection[T]) error

type SyncObjectHandler[T any, O any] func(systemge.Connection[T], O) (T, error)
type SyncObjectHandlers[T any, O any] map[string]SyncObjectHandler[T, O]

type AsyncObjectHandler[T any, O any] func(systemge.Connection[T], O)
type AsyncObjecthandlers[T any, O any] map[string]AsyncObjectHandler[T, O]

// executes the provided handlers in sequence.
// stops and returns an error if any handler returns an error.
// returns nil if all handlers succeed.
func NewAndHandler[T any](handlers ...HandlerWithError[T]) HandlerWithError[T] {
	return func(data T, connection systemge.Connection[T]) error {
		for _, handler := range handlers {
			if err := handler(data, connection); err != nil {
				return err
			}
		}
		return nil
	}
}

// executes the provided handlers in sequence.
// stops and returns nil if any handler returns nil.
// returns an error if all handlers return an error.
func NewOrHandler[T any](handlers ...HandlerWithError[T]) HandlerWithError[T] {
	return func(data T, connection systemge.Connection[T]) error {
		for _, handler := range handlers {
			if err := handler(data, connection); err != nil {
				return err
			}
		}
		return nil
	}
}

func NewConditionalHandler[T any](
	condition HandlerWithError[T],
	FailureHandler HandlerWithError[T],
	SuccessHandler HandlerWithError[T],
) HandlerWithError[T] {
	return func(data T, connection systemge.Connection[T]) error {
		if err := condition(data, connection); err != nil {
			return FailureHandler(data, connection)
		}
		return SuccessHandler(data, connection)
	}
}

func NewSuccessHandler[T any](
	condition HandlerWithError[T],
	SuccessHandler HandlerWithError[T],
) HandlerWithError[T] {
	return func(data T, connection systemge.Connection[T]) error {
		if err := condition(data, connection); err != nil {
			return err
		}
		return SuccessHandler(data, connection)
	}
}

func NewFailureHandler[T any](
	condition HandlerWithError[T],
	FailureHandler HandlerWithError[T],
) HandlerWithError[T] {
	return func(data T, connection systemge.Connection[T]) error {
		if err := condition(data, connection); err != nil {
			return FailureHandler(data, connection)
		}
		return nil
	}
}

// attempts to consume the provided amount of bytes from the token bucket rate limiter.
// returns an error if the rate limiter does not have enough tokens.
func NewByteRateLimitHandler(
	tokenBucketRateLimiterConfig *configs.TokenBucketRateLimiter,
) HandlerWithError[[]byte] {

	tokenBucketRateLimiter := tools.NewTokenBucketRateLimiter(tokenBucketRateLimiterConfig)
	return func(data []byte, connection systemge.Connection[[]byte]) error {
		tokenBucketRateLimiter.Consume(uint64(len(data)))
		return nil
	}
}

// attempts to consume 1 token from the token bucket rate limiter.
// returns an error if the rate limiter does not have enough tokens.
func NewMessageRateLimitHandler[T any](
	tokenBucketRateLimiterConfig *configs.TokenBucketRateLimiter,
) HandlerWithError[T] {

	tokenBucketRateLimiter := tools.NewTokenBucketRateLimiter(tokenBucketRateLimiterConfig)
	return func(data T, connection systemge.Connection[T]) error {
		tokenBucketRateLimiter.Consume(1)
		return nil
	}
}

// attempts to consume the provided amount of tokens from the token bucket rate limiter.
// returns an error if the rate limiter does not have enough tokens.
func NewCustomRateLimitHandler[T any](
	tokenBucketRateLimiterConfig *configs.TokenBucketRateLimiter,
	consumeFunc func(T, systemge.Connection[T]) uint64,
) HandlerWithError[T] {

	tokenBucketRateLimiter := tools.NewTokenBucketRateLimiter(tokenBucketRateLimiterConfig)
	return func(data T, connection systemge.Connection[T]) error {
		tokenBucketRateLimiter.Consume(consumeFunc(data, connection))
		return nil
	}
}

// executes the provided handler.
// if no error is returned, the provided response is written to the connection.
func NewResponseHandler[T any](
	getResponse func(T, systemge.Connection[T]) (T, error),
	writeTimeoutNs int64,
) HandlerWithError[T] {
	return func(incomingData T, connection systemge.Connection[T]) error {
		response, err := getResponse(incomingData, connection)
		if err != nil {
			return err
		}
		return connection.Write(response, writeTimeoutNs)
	}
}

type ObjectValidator[T any] func(T, systemge.Connection[T]) error

// executes the provided validator.
func NewValidationObjectHandler[T any](
	validator ObjectValidator[T],
) HandlerWithError[T] {
	return func(object T, connection systemge.Connection[T]) error {
		return validator(object, connection)
	}
}

type ObtainEnqueueConfigs[T any] func(T, systemge.Connection[T]) (token string, priority uint32, timeoutNs int64)

type queueWrapper[T any] struct {
	Object     T
	Connection systemge.Connection[T]
}

// obtains the token, priority and timeoutNs for the object and connection.
// enqueues the object and connection in the priority token queue.
func NewQueueHandler[T any](
	priorityTokenQueue *tools.PriorityTokenQueue[*queueWrapper[T]],
	obtainEnqueueConfigs ObtainEnqueueConfigs[T],
) HandlerWithError[T] {

	return func(object T, connection systemge.Connection[T]) error {
		token, priority, timeoutNs := obtainEnqueueConfigs(object, connection)
		queueWrapper := &queueWrapper[T]{
			object,
			connection,
		}
		return priorityTokenQueue.Push(token, queueWrapper, priority, timeoutNs)
	}
}

// repeatedly dequeues objects from the priority token queue and executes the provided handler.
func NewDequeueRoutine[T any](
	priorityTokenQueue *tools.PriorityTokenQueue[*queueWrapper[T]],
	readHandler HandlerWithError[T],
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

type objectHandlerWrapper[T any, O any] struct {
	object     O
	connection systemge.Connection[T]
}

// creates a new topic manager with the provided handlers.
func NewObjectTopicManager[T any, O any](
	asyncObjectHandlers AsyncObjecthandlers[T, O],
	syncObjectHandlers SyncObjectHandlers[T, O],
	unknownAsyncObjectHandler AsyncObjectHandler[T, O],
	unknownSyncObjectHandler SyncObjectHandler[T, O],
	topicManagerConfig *configs.TopicManager,
) (*tools.TopicManager[objectHandlerWrapper[T, O]], error) {

	topicHandlers := tools.TopicHandlers[objectHandlerWrapper[T, O]]{}
	for topic, handler := range asyncObjectHandlers {
		topicHandlers[topic] = func(mhw objectHandlerWrapper[T, O]) {
			handler(mhw.connection, mhw.object)
		}
	}
	for topic, handler := range syncObjectHandlers {
		topicHandlers[topic] = func(mhw objectHandlerWrapper[T, O]) {
			response, err := handler(mhw.connection, mhw.object)
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

	var unknownTopicHandler tools.TopicHandler[objectHandlerWrapper[T, O]]
	if unknownAsyncObjectHandler != nil || unknownSyncObjectHandler != nil {
		unknownTopicHandler = func(mhw objectHandlerWrapper[T, O]) {
			if unknownAsyncObjectHandler != nil {
				unknownAsyncObjectHandler(mhw.connection, mhw.object)
			}
			if unknownSyncObjectHandler != nil {
				response, err := unknownSyncObjectHandler(mhw.connection, mhw.object)
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

	return tools.NewTopicManager(
		topicManagerConfig,
		topicHandlers,
		unknownTopicHandler,
	)
}

// retrieves the topic and object from the provided data and connection and executes the handler.
func NewTopicHandler[T any, O any](
	topicManager *tools.TopicManager[objectHandlerWrapper[T, O]],
	retrieveTopicAndObject func(T, systemge.Connection[T]) (string, O, error), // returns topic and O since returning topic and T would be inefficient/redundant in most scenarios (repeated de/serialization) (O may == T anyway)
) HandlerWithError[T] {

	return func(data T, connection systemge.Connection[T]) error {
		topic, object, err := retrieveTopicAndObject(data, connection)
		if err != nil {
			return errors.New("could not retrieve object")
		}
		return topicManager.Handle(
			topic,
			objectHandlerWrapper[T, O]{
				object,
				connection,
			},
		)
	}
}
