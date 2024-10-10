package Tools

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
)

type ObjectHandler func(object any, args ...any) error
type ObjectDeserializer func([]byte) (any, error)
type ObjectValidator func(any) error

type ReceptionHandler func([]byte) error
type ReceptionHandlerFactory struct {
	ByteRateLimiterConfig    *Config.TokenBucketRateLimiter
	MessageRateLimiterConfig *Config.TokenBucketRateLimiter
	ObjectDeserializer       ObjectDeserializer
	ObjectValidator          ObjectValidator
	ObjectHandler            ObjectHandler
}

func NewReceptionHandlerStruct(byteRateLimiterConfig *Config.TokenBucketRateLimiter, messageRateLimiterConfig *Config.TokenBucketRateLimiter, objectDeserializer ObjectDeserializer, objectValidator ObjectValidator, objectHandler ObjectHandler) *ReceptionHandlerFactory {
	return &ReceptionHandlerFactory{
		ByteRateLimiterConfig:    byteRateLimiterConfig,
		MessageRateLimiterConfig: messageRateLimiterConfig,
		ObjectDeserializer:       objectDeserializer,
		ObjectValidator:          objectValidator,
		ObjectHandler:            objectHandler,
	}
}

func (receptionHandlerStruct *ReceptionHandlerFactory) NewReceptionHandler(args any) ReceptionHandler {
	var byteRateLimiter *TokenBucketRateLimiter
	if receptionHandlerStruct.ByteRateLimiterConfig != nil {
		byteRateLimiter = NewTokenBucketRateLimiter(receptionHandlerStruct.ByteRateLimiterConfig)
	}
	var messageRateLimiter *TokenBucketRateLimiter
	if receptionHandlerStruct.MessageRateLimiterConfig != nil {
		messageRateLimiter = NewTokenBucketRateLimiter(receptionHandlerStruct.MessageRateLimiterConfig)
	}
	return func(bytes []byte) error {
		if byteRateLimiter != nil && !byteRateLimiter.Consume(uint64(len(bytes))) {
			if websocketServer.GetEventHandler() != nil {
				if event := websocketServer.GetEventHandler().Handle(Event.New(
					Event.RateLimited,
					Event.Context{
						Event.SessionId:       sessionId,
						Event.Identity:        identity,
						Event.Address:         websocketClient.GetAddress(),
						Event.RateLimiterType: Event.TokenBucket,
						Event.TokenBucketType: Event.Bytes,
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

		if messageRateLimiter != nil && !messageRateLimiter.Consume(1) {
			if websocketServer.GetEventHandler() != nil {
				if event := websocketServer.GetEventHandler().Handle(Event.New(
					Event.RateLimited,
					Event.Context{
						Event.SessionId:       sessionId,
						Event.Identity:        identity,
						Event.Address:         websocketClient.GetAddress(),
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

		object, err := receptionHandlerStruct.ObjectDeserializer(bytes)
		if err != nil {
			if websocketServer.GetEventHandler() != nil {
				websocketServer.GetEventHandler().Handle(Event.New(
					Event.DeserializingFailed,
					Event.Context{
						Event.SessionId: sessionId,
						Event.Identity:  identity,
						Event.Address:   websocketClient.GetAddress(),
						Event.Bytes:     string(bytes),
						Event.Error:     err.Error(),
					},
					Event.Skip,
				))
			}
			return errors.New(Event.DeserializingFailed)
		}

		if err := receptionHandlerStruct.ObjectValidator(object); err != nil {
			if websocketServer.GetEventHandler() != nil {
				event := websocketServer.GetEventHandler().Handle(Event.New(
					Event.InvalidMessage,
					Event.Context{
						Event.SessionId: sessionId,
						Event.Identity:  identity,
						Event.Address:   websocketClient.GetAddress(),
						Event.Bytes:     string(bytes),
						Event.Error:     err.Error(),
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

		return receptionHandlerStruct.ObjectHandler(object, websocketServer, websocketClient, identity, sessionId)
	}
}
