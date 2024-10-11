package ReceptionHandler

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
)

type ReceptionHandler func([]byte) error

type ByteHandler[T any] func([]byte) (T, error)
type ObjectDeserializer[T any] func([]byte) (T, error)

type ObjectHandler[T any] func(T) error
type ObjectValidator[T any] func(T) error

type ObtainResponseToken[T any] func(T) string
type ObtainEnqueueConfigs[T any] func(T) (string, uint32, uint32)

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

func NewValidationReceptionHandler[T any](
	byteHandler ByteHandler[T],
	objectHandler ObjectHandler[T],

) ReceptionHandler {
	return func(bytes []byte) error {

		/* if byteRateLimiter != nil && !byteRateLimiter.Consume(uint64(len(bytes))) {
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
		} */

		object, err := byteHandler(bytes)
		if err != nil {
			return errors.New(Event.DeserializingFailed)
		}

		return objectHandler(object)
	}
}
