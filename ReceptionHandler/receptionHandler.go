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

type AssertIsResponse[T any] func(T) bool
type ObtainResponseToken[T any] func(T) string
type ObtainEnqueueConfigs[T any] func(T) (string, uint32, uint32)

func NewValidationReceptionHandler[T any](
	eventHandler *Event.Handler,
	defaultContext Event.Context,
	byteRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageRateLimiterConfig *Config.TokenBucketRateLimiter,

	deserializer ObjectDeserializer[T],
	validator ObjectValidator[T],

	requestResponseManager *Tools.RequestResponseManager[T],
	priorityTokenQueue *Tools.PriorityTokenQueue[T],

	getObjectIsResponse AssertIsResponse[T],
	getObjectResponseToken ObtainResponseToken[T],
	getObjectQueueConfig ObtainEnqueueConfigs[T],
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

		if err := validator(object); err != nil {
			if eventHandler != nil {
				event := eventHandler.Handle(Event.New(
					Event.InvalidMessage,
					Event.Context{
						Event.Bytes: string(bytes),
						Event.Error: err.Error(),
					}.Merge(defaultContext),
					Event.Skip,
					Event.Continue,
				))
				if event.GetAction() == Event.Skip {
					return errors.New(Event.InvalidMessage)
				}
			} else {
				return errors.New(Event.InvalidMessage)
			}
		}

		if getObjectIsResponse(object) {
			if requestResponseManager != nil {
				if err := requestResponseManager.AddResponse(getObjectResponseToken(object), object); err != nil {
					return err
				}
				return nil
			}
		}

		token, priority, timeoutMs := getObjectQueueConfig(object)
		return priorityTokenQueue.Push(token, object, priority, timeoutMs)
	}
}
