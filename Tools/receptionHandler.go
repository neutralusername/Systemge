package Tools

import (
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Status"
)

type ReceptionHandler[S any] struct {
	onStart     OnReceptionHandlerStart[S]
	onStop      OnReceptionHandlerStop[S]
	onReception OnReception[S]
	status      int
	statusMutex sync.RWMutex
}

func NewReceptionHandler[S any](
	onStart OnReceptionHandlerStart[S],
	onStop OnReceptionHandlerStop[S],
	OnReception OnReception[S],
) *ReceptionHandler[S] {
	return &ReceptionHandler[S]{
		onStart:     onStart,
		onStop:      onStop,
		onReception: OnReception,
		status:      Status.Stopped,
	}
}

func (handler *ReceptionHandler[S]) HandleReception(bytes []byte, structName123 S) error {
	handler.statusMutex.RLock()
	defer handler.statusMutex.RUnlock()
	if handler.status == Status.Stopped {
		return errors.New("handler is stopped")
	}
	if handler.onReception == nil {
		return errors.New("onHandle is nil")
	}
	return handler.onReception(bytes, structName123)
}

func (handler *ReceptionHandler[S]) Start(structName123 S) error {
	handler.statusMutex.Lock()
	defer handler.statusMutex.Unlock()
	if handler.status != Status.Stopped {
		return errors.New("handler is not stopped")
	}
	handler.status = Status.Pending
	if handler.onStart != nil {
		err := handler.onStart(structName123)
		if err != nil {
			handler.status = Status.Stopped
			return err
		}
	}
	handler.status = Status.Started
	return nil
}

func (handler *ReceptionHandler[S]) Stop(structName123 S) error {
	handler.statusMutex.Lock()
	defer handler.statusMutex.Unlock()
	if handler.status != Status.Started {
		return errors.New("handler is not started")
	}
	handler.status = Status.Pending
	if handler.onStop != nil {
		err := handler.onStop(structName123)
		if err != nil {
			handler.status = Status.Started
			return err
		}
	}
	handler.status = Status.Stopped
	return nil
}

func (handler *ReceptionHandler[S]) GetStatus() int {
	return handler.status
}

type ReceptionHandlerFactory[S any] func() OnReception[S]

type OnReception[S any] func([]byte, S) error

type OnReceptionHandlerStart[S any] func(S) error
type OnReceptionHandlerStop[S any] func(S) error

type ByteHandler[S any] func([]byte, S) error
type ObjectDeserializer[T any, S any] func([]byte, S) (T, error)
type ObjectHandler[T any, S any] func(T, S) error

func NewReceptionHandlerFactory[T any, S any](
	byteHandler ByteHandler[S],
	deserializer ObjectDeserializer[T, S],
	objectHandler ObjectHandler[T, S],
) ReceptionHandlerFactory[S] {
	return func() OnReception[S] {
		return NewOnReception[T, S](
			byteHandler,
			deserializer,
			objectHandler,
		)
	}
}

func NewOnReception[T any, S any](
	byteHandler ByteHandler[S],
	deserializer ObjectDeserializer[T, S],
	objectHandler ObjectHandler[T, S],
) OnReception[S] {
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

func NewValidationObjectHandler[T any, S any](validator ObjectValidator[T, S]) ObjectHandler[T, S] {
	return func(object T, structName123 S) error {
		return validator(object, structName123)
	}
}

/* // resultHandler requires check for nil if applicable
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
} */

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
