package ReceptionHandler

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
)

type ReceptionHandler func([]byte) error

type ObjectDeserializer[T any] func([]byte) (T, error)
type ObjectValidator[T any] func(T) error

type ObtainResponseToken[T any] func(T) string
type ObtainEnqueueConfigs[T any] func(T) (string, uint32, uint32)

type ObjectHandler[T any] func(T) error

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
				if err := requestResponseManager.AddResponse(responseToken, object); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func NewValidationObjectHandler[T any](
	validator ObjectValidator[T],
) ObjectHandler[T] {
	return func(object T) error {
		if err := validator(object); err != nil {
			return err
		}
		return nil
	}
}

func ChainObjectHandlers[T any](handlers ...ObjectHandler[T]) ObjectHandler[T] {
	return func(object T) error {
		for _, handler := range handlers {
			if err := handler(object); err != nil {
				return err
			}
		}
		return nil
	}
}

func NewValidationReceptionHandler[T any](
	eventHandler *Event.Handler,
	defaultContext Event.Context,
	byteRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageRateLimiterConfig *Config.TokenBucketRateLimiter,

	deserializer ObjectDeserializer[T],
	objectHandler ObjectHandler[T],

) ReceptionHandler {

	var byteRateLimiter *Tools.TokenBucketRateLimiter
	if byteRateLimiterConfig != nil {
		byteRateLimiter = Tools.NewTokenBucketRateLimiter(byteRateLimiterConfig)
	}
	var messageRateLimiter *Tools.TokenBucketRateLimiter
	if messageRateLimiterConfig != nil {
		messageRateLimiter = Tools.NewTokenBucketRateLimiter(messageRateLimiterConfig)
	}

	return func(bytes []byte) error {

		if byteRateLimiter != nil && !byteRateLimiter.Consume(uint64(len(bytes))) {
			if eventHandler != nil {
				if event := eventHandler.Handle(Event.New(
					Event.RateLimited,
					Event.Context{
						Event.RateLimiterType: Event.TokenBucket,
						Event.TokenBucketType: Event.Bytes,
						Event.Bytes:           string(bytes),
					}.Merge(defaultContext),
					Event.Skip,
					Event.Continue,
				)); event.GetAction() == Event.Skip {
					return errors.New(Event.RateLimited)
				}
			} else {
				return errors.New(Event.RateLimited)
			}
		}

		if messageRateLimiter != nil && !messageRateLimiter.Consume(1) {
			if eventHandler != nil {
				if event := eventHandler.Handle(Event.New(
					Event.RateLimited,
					Event.Context{
						Event.RateLimiterType: Event.TokenBucket,
						Event.TokenBucketType: Event.Messages,
						Event.Bytes:           string(bytes),
					}.Merge(defaultContext),
					Event.Skip,
					Event.Continue,
				)); event.GetAction() == Event.Skip {
					return errors.New(Event.RateLimited)
				}
			} else {
				return errors.New(Event.RateLimited)
			}
		}

		object, err := deserializer(bytes)
		if err != nil {
			if eventHandler != nil {
				eventHandler.Handle(Event.New(
					Event.DeserializingFailed,
					Event.Context{
						Event.Bytes: string(bytes),
						Event.Error: err.Error(),
					}.Merge(defaultContext),
					Event.Skip,
				))
			}
			return errors.New(Event.DeserializingFailed)
		}

		return objectHandler(object)
	}
}
