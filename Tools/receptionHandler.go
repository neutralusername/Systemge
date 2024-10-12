package Tools

import (
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Status"
)

type ReceptionHandler[C any] struct {
	onStart     OnReceptionHandlerStart[C]
	onStop      OnReceptionHandlerStop[C]
	onReception OnReception[C]
	status      int
	statusMutex sync.RWMutex
}

func NewReceptionHandler[C any](
	onStart OnReceptionHandlerStart[C],
	onStop OnReceptionHandlerStop[C],
	OnReception OnReception[C],
) *ReceptionHandler[C] {
	return &ReceptionHandler[C]{
		onStart:     onStart,
		onStop:      onStop,
		onReception: OnReception,
		status:      Status.Stopped,
	}
}

func (handler *ReceptionHandler[C]) HandleReception(bytes []byte, caller C) error {
	handler.statusMutex.RLock()
	defer handler.statusMutex.RUnlock()

	if handler.status == Status.Stopped {
		return errors.New("handler is stopped")
	}
	if handler.onReception == nil {
		return errors.New("onHandle is nil")
	}
	return handler.onReception(bytes, caller)
}

func (handler *ReceptionHandler[C]) Start(caller C) error {
	handler.statusMutex.Lock()
	defer handler.statusMutex.Unlock()

	if handler.status != Status.Stopped {
		return errors.New("handler is not stopped")
	}
	handler.status = Status.Pending
	if handler.onStart != nil {
		err := handler.onStart(caller)
		if err != nil {
			handler.status = Status.Stopped
			return err
		}
	}
	handler.status = Status.Started
	return nil
}

func (handler *ReceptionHandler[C]) Stop(caller C) error {
	handler.statusMutex.Lock()
	defer handler.statusMutex.Unlock()

	if handler.status != Status.Started {
		return errors.New("handler is not started")
	}
	handler.status = Status.Pending
	if handler.onStop != nil {
		err := handler.onStop(caller)
		if err != nil {
			handler.status = Status.Started
			return err
		}
	}
	handler.status = Status.Stopped
	return nil
}

func (handler *ReceptionHandler[C]) GetStatus(lock bool) int {
	if lock {
		handler.statusMutex.RLock()
		defer handler.statusMutex.RUnlock()
	}
	return handler.status
}

type ReceptionHandlerFactory[C any] func() *ReceptionHandler[C]

type OnReceptionHandlerStart[C any] func(C) error
type OnReceptionHandlerStop[C any] func(C) error

type OnReception[C any] func([]byte, C) error

type ByteHandler[C any] func([]byte, C) error
type ObjectDeserializer[T any, C any] func([]byte, C) (T, error)
type ObjectHandler[T any, C any] func(T, C) error

func NewReceptionHandlerFactory[C any](
	onStart OnReceptionHandlerStart[C],
	onStop OnReceptionHandlerStop[C],
	onReception OnReception[C],
) ReceptionHandlerFactory[C] {
	return func() *ReceptionHandler[C] {
		return NewReceptionHandler[C](onStart, onStop, onReception)
	}
}

func NewOnReception[T any, C any](
	byteHandler ByteHandler[C],
	deserializer ObjectDeserializer[T, C],
	objectHandler ObjectHandler[T, C],
) OnReception[C] {
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

// executes all handlers in order, return error if any handler returns an error
func NewChainObjecthandler[T any, C any](handlers ...ObjectHandler[T, C]) ObjectHandler[T, C] {
	return func(object T, caller C) error {
		for _, handler := range handlers {
			if err := handler(object, caller); err != nil {
				return err
			}
		}
		return nil
	}
}

type ObtainEnqueueConfigs[T any, C any] func(T, C) (token string, priority uint32, timeout uint32)

func NewQueueObjectHandler[T any, C any](
	priorityTokenQueue *PriorityTokenQueue[T],
	obtainEnqueueConfigs ObtainEnqueueConfigs[T, C],
) ObjectHandler[T, C] {
	return func(object T, caller C) error {
		token, priority, timeoutMs := obtainEnqueueConfigs(object, caller)
		return priorityTokenQueue.Push(token, object, priority, timeoutMs)
	}
}

type ObtainResponseToken[T any, C any] func(T, C) string

func NewResponseObjectHandler[T any, C any](
	requestResponseManager *RequestResponseManager[T],
	obtainResponseToken ObtainResponseToken[T, C],
) ObjectHandler[T, C] {
	return func(object T, caller C) error {
		responseToken := obtainResponseToken(object, caller)
		if responseToken != "" {
			if requestResponseManager != nil {
				return requestResponseManager.AddResponse(responseToken, object)
			}
		}
		return nil
	}
}

type ObjectValidator[T any, C any] func(T, C) error

func NewValidationObjectHandler[T any, C any](validator ObjectValidator[T, C]) ObjectHandler[T, C] {
	return func(object T, caller C) error {
		return validator(object, caller)
	}
}

type ObtainTopic[T any, C any] func(T, C) string
type ResultHandler[T any, R any, C any] func(T, R, C) error

// resultHandler requires check for nil if applicable
func NewTopicObjectHandler[T any, R any, C any](
	topicManager *TopicManager[T, R],
	obtainTopic func(T) string,
	resultHandler ResultHandler[T, R, C],
) ObjectHandler[T, C] {
	return func(object T, caller C) error {
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
