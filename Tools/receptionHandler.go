package Tools

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
)

type ObjectHandler func(object any) error
type ObjectDeserializer func([]byte) (any, error)
type ObjectValidator func(any) error

type ReceptionHandler struct {
	byteRateLimiter    *TokenBucketRateLimiter
	messageRateLimiter *TokenBucketRateLimiter
	objectDeserializer ObjectDeserializer
	objectValidator    ObjectValidator
	objectHandler      ObjectHandler

	getEventHandler              func() *Event.Handler
	receptionHandlerEventContext Event.Context
}

func NewReceptionHandler(byteRateLimiterConfig *Config.TokenBucketRateLimiter, messageRateLimiterConfig *Config.TokenBucketRateLimiter, objectDeserializer ObjectDeserializer, objectValidator ObjectValidator, objectHandler ObjectHandler) *ReceptionHandler {
	receptionHandler := &ReceptionHandler{
		objectDeserializer: objectDeserializer,
		objectValidator:    objectValidator,
		objectHandler:      objectHandler,
	}
	if byteRateLimiterConfig != nil {
		receptionHandler.byteRateLimiter = NewTokenBucketRateLimiter(byteRateLimiterConfig)
	}
	if messageRateLimiterConfig != nil {
		receptionHandler.messageRateLimiter = NewTokenBucketRateLimiter(messageRateLimiterConfig)
	}
	return receptionHandler
}

func (receptionHandler *ReceptionHandler) Handle(bytes []byte) error {
	if receptionHandler.byteRateLimiter != nil && !receptionHandler.byteRateLimiter.Consume(uint64(len(bytes))) {
		if receptionHandler.getEventHandler() != nil {
			if event := receptionHandler.getEventHandler().Handle(Event.New(
				Event.RateLimited,
				Event.Context{
					Event.RateLimiterType: Event.TokenBucket,
					Event.TokenBucketType: Event.Bytes,
					Event.Bytes:           string(bytes),
				}.Merge(receptionHandler.receptionHandlerEventContext),
				Event.Skip,
				Event.Continue,
			)); event.GetAction() == Event.Skip {
				return errors.New(Event.RateLimited)
			}
		} else {
			return errors.New(Event.RateLimited)
		}
	}

	if receptionHandler.messageRateLimiter != nil && !receptionHandler.messageRateLimiter.Consume(1) {
		if receptionHandler.getEventHandler() != nil {
			if event := receptionHandler.getEventHandler().Handle(Event.New(
				Event.RateLimited,
				Event.Context{
					Event.RateLimiterType: Event.TokenBucket,
					Event.TokenBucketType: Event.Messages,
					Event.Bytes:           string(bytes),
				},
				Event.Skip,
				Event.Continue,
			)); event.GetAction() == Event.Skip {
				return errors.New(Event.RateLimited)
			}
		} else {
			return errors.New(Event.RateLimited)
		}
	}

	object, err := receptionHandler.objectDeserializer(bytes)
	if err != nil {
		if receptionHandler.getEventHandler() != nil {
			receptionHandler.getEventHandler().Handle(Event.New(
				Event.DeserializingFailed,
				Event.Context{
					Event.Bytes: string(bytes),
					Event.Error: err.Error(),
				},
				Event.Skip,
			))
		}
		return errors.New(Event.DeserializingFailed)
	}

	if err := receptionHandler.objectValidator(object); err != nil {
		if receptionHandler.getEventHandler() != nil {
			event := receptionHandler.getEventHandler().Handle(Event.New(
				Event.InvalidMessage,
				Event.Context{
					Event.Bytes: string(bytes),
					Event.Error: err.Error(),
				},
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

	return receptionHandler.objectHandler(object)
}
