package Pipeline

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
)

type Pipeline struct {
	name             string
	rateLimiterBytes *Tools.TokenBucketRateLimiter
	rateLimiterCalls *Tools.TokenBucketRateLimiter
	deserializer     func([]byte) (any, error)
	validator        func(any) error

	eventHandler Event.Handler
}

func NewPipeline(name string, rateLimiterBytes *Tools.TokenBucketRateLimiter, rateLimiterCalls *Tools.TokenBucketRateLimiter, deserializer func([]byte) (any, error), validator func(any) error, eventHandler Event.Handler) (*Pipeline, error) {
	if deserializer == nil {
		return nil, errors.New("deserializer is required")
	}
	return &Pipeline{
		name:             name,
		rateLimiterBytes: rateLimiterBytes,
		rateLimiterCalls: rateLimiterCalls,
		deserializer:     deserializer,
		validator:        validator,
		eventHandler:     eventHandler,
	}, nil
}

func (pipeline *Pipeline) DoStuff(bytes []byte) (any, error) {
	if pipeline.rateLimiterBytes != nil && !pipeline.rateLimiterBytes.Consume(uint64(len(bytes))) {
		return nil, errors.New("byte rate limit exceeded")
	}
	if pipeline.rateLimiterCalls != nil && !pipeline.rateLimiterCalls.Consume(1) {
		return nil, errors.New("call rate limit exceeded")
	}
	data, err := pipeline.deserializer(bytes)
	if err != nil {
		return nil, err
	}
	if pipeline.validator != nil {
		err = pipeline.validator(data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (server *Pipeline) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(server.GetContext())
	if server.eventHandler != nil {
		server.eventHandler(event)
	}
	return event
}
func (server *Pipeline) GetContext() Event.Context {
	return Event.Context{
		Event.ServiceType: Event.Pipeline,
		Event.ServiceName: server.name,
		Event.Function:    Event.GetCallerFuncName(2),
	}
}
