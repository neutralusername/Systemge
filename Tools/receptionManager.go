package Tools

import (
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Status"
)

type ReceptionManager[C any] struct {
	onStart     OnReceptionManagerStart[C]
	onStop      OnReceptionManagerStop[C]
	onReception OnReceptionManagerHandle[C]
	status      int
	statusMutex sync.RWMutex
}

func NewReceptionManager[C any](
	onStart OnReceptionManagerStart[C],
	onStop OnReceptionManagerStop[C],
	OnReception OnReceptionManagerHandle[C],
) *ReceptionManager[C] {
	return &ReceptionManager[C]{
		onStart:     onStart,
		onStop:      onStop,
		onReception: OnReception,
		status:      Status.Stopped,
	}
}

func (manager *ReceptionManager[C]) Handle(bytes []byte, caller C) error {
	manager.statusMutex.RLock()
	defer manager.statusMutex.RUnlock()

	if manager.status == Status.Stopped {
		return errors.New("handler is stopped")
	}
	if manager.onReception == nil {
		return errors.New("onHandle is nil")
	}
	return manager.onReception(bytes, caller)
}

func (handler *ReceptionManager[C]) Start(caller C) error {
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

func (handler *ReceptionManager[C]) Stop(caller C) error {
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

func (handler *ReceptionManager[C]) GetStatus(lock bool) int {
	if lock {
		handler.statusMutex.RLock()
		defer handler.statusMutex.RUnlock()
	}
	return handler.status
}

type ReceptionManagerFactory[C any] func() *ReceptionManager[C]

type OnReceptionManagerStart[C any] func(C) error
type OnReceptionManagerStop[C any] func(C) error

type OnReceptionManagerHandle[C any] func([]byte, C) error

type ByteHandler[C any] func([]byte, C) error
type ObjectDeserializer[O any, C any] func([]byte, C) (O, error)
type ObjectHandler[O any, C any] func(O, C) error

func NewReceptionManagerFactory[C any](
	onStart OnReceptionManagerStart[C],
	onStop OnReceptionManagerStop[C],
	onReception OnReceptionManagerHandle[C],
) ReceptionManagerFactory[C] {
	return func() *ReceptionManager[C] {
		return NewReceptionManager[C](onStart, onStop, onReception)
	}
}

func NewOnReceptionManagerHandle[O any, C any](
	byteHandler ByteHandler[C],
	deserializer ObjectDeserializer[O, C],
	objectHandler ObjectHandler[O, C],
) OnReceptionManagerHandle[C] {
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

	requestResponseManager *RequestResponseManager[O],
	obtainResponseToken ObtainResponseToken[O, C],

	priorityQueue *PriorityTokenQueue[O],
	obtainEnqueueConfigs ObtainEnqueueConfigs[O, C],

	// topicManager *Tools.TopicManager,
) ReceptionManagerFactory[C] {

	byteHandlers := []ByteHandler[C]{}
	if byteRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, NewByteRateLimitByteHandler[C](NewTokenBucketRateLimiter(byteRateLimiterConfig)))
	}
	if messageRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, NewMessageRateLimitByteHandler[C](NewTokenBucketRateLimiter(messageRateLimiterConfig)))
	}

	objectHandlers := []ObjectHandler[O, C]{}
	if messageValidator != nil {
		objectHandlers = append(objectHandlers, messageValidator)
	}
	if requestResponseManager != nil && obtainResponseToken != nil {
		objectHandlers = append(objectHandlers, NewResponseObjectHandler(requestResponseManager, obtainResponseToken))
	}
	if priorityQueue != nil && obtainEnqueueConfigs != nil {
		objectHandlers = append(objectHandlers, NewQueueObjectHandler(priorityQueue, obtainEnqueueConfigs))
	}

	return NewReceptionManagerFactory(
		nil,
		nil,
		NewOnReceptionManagerHandle(
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

type ObtainResponseToken[O any, C any] func(O, C) string

func NewResponseObjectHandler[O any, C any](
	requestResponseManager *RequestResponseManager[O],
	obtainResponseToken ObtainResponseToken[O, C],
) ObjectHandler[O, C] {
	return func(object O, caller C) error {
		responseToken := obtainResponseToken(object, caller)
		if responseToken != "" {
			if requestResponseManager != nil {
				return requestResponseManager.AddResponse(responseToken, object)
			}
		}
		return nil
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
