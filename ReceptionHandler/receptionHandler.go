package ReceptionHandler

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
)

type ReceptionHandler struct {
	name             string
	rateLimiterBytes *Tools.TokenBucketRateLimiter
	rateLimiterCalls *Tools.TokenBucketRateLimiter
	deserializer     func([]byte, ...any) (any, error)
	validator        func(any, ...any) error

	eventHandler Event.Handler
}

type Validator func(any, ...any) error
type Deserializer func([]byte, ...any) (any, error)

func NewReceptionHandler(name string, rateLimiterBytes *Tools.TokenBucketRateLimiter, rateLimiterCalls *Tools.TokenBucketRateLimiter, deserializer Deserializer, validator Validator, eventHandler Event.Handler) (*ReceptionHandler, error) {
	if deserializer == nil {
		return nil, errors.New("deserializer is required")
	}
	return &ReceptionHandler{
		name:             name,
		rateLimiterBytes: rateLimiterBytes,
		rateLimiterCalls: rateLimiterCalls,
		deserializer:     deserializer,
		validator:        validator,
		eventHandler:     eventHandler,
	}, nil
}

func (handler *ReceptionHandler) HandleReception(bytes []byte, args ...any) (any, error) {
	if handler.rateLimiterBytes != nil && !handler.rateLimiterBytes.Consume(uint64(len(bytes))) {
		return nil, errors.New("byte rate limit exceeded")
	}
	if handler.rateLimiterCalls != nil && !handler.rateLimiterCalls.Consume(1) {
		return nil, errors.New("call rate limit exceeded")
	}
	data, err := handler.deserializer(bytes, args...)
	if err != nil {
		return nil, err
	}
	if handler.validator != nil {
		err = handler.validator(data, args...)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (handler *ReceptionHandler) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(handler.GetContext())
	if handler.eventHandler != nil {
		handler.eventHandler(event)
	}
	return event
}
func (handler *ReceptionHandler) GetContext() Event.Context {
	return Event.Context{
		Event.ServiceType: Event.Pipeline,
		Event.ServiceName: handler.name,
		Event.Function:    Event.GetCallerFuncName(2),
	}
}
