package WebsocketServer

import (
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

type ReceptionHandler func([]byte) error
type ReceptionHandlerFactory func(websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) ReceptionHandler

type ObjectHandler func(object any, websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error
type ObjectDeserializer func([]byte) (any, error)
type ObjectValidator func(any) error

type InitializerFunc func(websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) any

func NewWebsocketTopicManager(config *Config.TopicManager, topicObjectHandlers map[string]ObjectHandler, unknownObjectHandler ObjectHandler) *Tools.TopicManager {
	topicHandlers := make(Tools.TopicHandlers)
	for topic, objectHandler := range topicObjectHandlers {
		topicHandlers[topic] = func(args ...any) (any, error) {
			message := args[0].(*Message.Message)
			websocketServer := args[1].(*WebsocketServer)
			websocketClient := args[2].(*WebsocketClient.WebsocketClient)
			identity := args[3].(string)
			sessionId := args[4].(string)
			return nil, objectHandler(message, websocketServer, websocketClient, identity, sessionId)
		}
	}
	unknownTopicHandler := func(args ...any) (any, error) {
		message := args[0].(*Message.Message)
		websocketServer := args[1].(*WebsocketServer)
		websocketClient := args[2].(*WebsocketClient.WebsocketClient)
		identity := args[3].(string)
		sessionId := args[4].(string)
		return nil, unknownObjectHandler(message, websocketServer, websocketClient, identity, sessionId)
	}
	return Tools.NewTopicManager(config, topicHandlers, unknownTopicHandler)
}

func NewDefaultReceptionHandlerFactory() ReceptionHandlerFactory {
	return func(websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) ReceptionHandler {
		return func(bytes []byte) error {
			return nil
		}
	}
}

func NewValidationMessageReceptionHandlerFactory(byteRateLimiterConfig *Config.TokenBucketRateLimiter, messageRateLimiterConfig *Config.TokenBucketRateLimiter, messageValidatorConfig *Config.MessageValidator, topicManager *Tools.TopicManager, priorityQueue *Tools.PriorityTokenQueue[*Message.Message], topicPriorities map[string]uint32, topicTimeoutMs map[string]uint64) ReceptionHandlerFactory {
	objectDeserializer := func(messageBytes []byte) (any, error) {
		return Message.Deserialize(messageBytes)
	}
	objectValidator := func(object any) error {
		message := object.(*Message.Message)
		if messageValidatorConfig.MinSyncTokenSize >= 0 && len(message.GetSyncToken()) < messageValidatorConfig.MinSyncTokenSize {
			return errors.New("message contains sync token")
		}
		if messageValidatorConfig.MaxSyncTokenSize >= 0 && len(message.GetSyncToken()) > messageValidatorConfig.MaxSyncTokenSize {
			return errors.New("message contains sync token")
		}
		if messageValidatorConfig.MinTopicSize >= 0 && len(message.GetTopic()) < messageValidatorConfig.MinTopicSize {
			return errors.New("message missing topic")
		}
		if messageValidatorConfig.MaxTopicSize >= 0 && len(message.GetTopic()) > messageValidatorConfig.MaxTopicSize {
			return errors.New("message missing topic")
		}
		if messageValidatorConfig.MinPayloadSize >= 0 && len(message.GetPayload()) < messageValidatorConfig.MinPayloadSize {
			return errors.New("message payload exceeds maximum size")
		}
		if messageValidatorConfig.MaxPayloadSize >= 0 && len(message.GetPayload()) > messageValidatorConfig.MaxPayloadSize {
			return errors.New("message payload exceeds maximum size")
		}
		return nil
	}

	handleTopic := func(message *Message.Message, websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error {
		response, err := topicManager.HandleTopic(message.GetTopic(), message, websocketServer, websocketClient, identity, sessionId)
		if err != nil {
			// event
			return err
		}

		if response != nil {
			// handle response
		}
		return nil
	}
	var initializerFunc InitializerFunc
	var messageHandler func(*Message.Message, *WebsocketServer, *WebsocketClient.WebsocketClient, string, string) error
	var objectHandler ObjectHandler = func(object any, websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error {
		message := object.(*Message.Message)
		if message.IsResponse() {
			websocketServer.requestResponseManager.AddResponse(message.GetSyncToken(), message) // can't be accessed by custom functions outside of this package currently
			return nil
		}
		return messageHandler(message, websocketServer, websocketClient, identity, sessionId)
	}
	if priorityQueue != nil {
		if handleTopic != nil {
			initializerFunc = func(websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) any {
				go func() {
					for {
						select {
						case message := <-priorityQueue.PopChannel():
							handleTopic(message, websocketServer, websocketClient, identity, sessionId)
						case <-websocketClient.GetCloseChannel():
							if priorityQueue.Len() == 0 {
								return
							}
						}
					}
				}()
				return nil
			}
		}
		mutex := &sync.Mutex{}
		messageHandler = func(message *Message.Message, websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error {
			mutex.Lock()
			priority := topicPriorities[message.GetTopic()]
			timeoutMs := topicTimeoutMs[message.GetTopic()]
			mutex.Unlock()
			return priorityQueue.Push("", message, priority, timeoutMs)
		}
	} else {
		messageHandler = func(message *Message.Message, websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error {
			return handleTopic(message, websocketServer, websocketClient, identity, sessionId)
		}
	}
	return NewValidationReceptionHandlerFactory(byteRateLimiterConfig, messageRateLimiterConfig, objectDeserializer, objectValidator, objectHandler, initializerFunc)
}

func NewValidationReceptionHandlerFactory(byteRateLimiterConfig *Config.TokenBucketRateLimiter, messageRateLimiterConfig *Config.TokenBucketRateLimiter, deserializer ObjectDeserializer, validator ObjectValidator, objectHandler ObjectHandler, initializerFunc InitializerFunc) ReceptionHandlerFactory {

	return func(websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) ReceptionHandler {
		var byteRateLimiter *Tools.TokenBucketRateLimiter
		if byteRateLimiterConfig != nil {
			byteRateLimiter = Tools.NewTokenBucketRateLimiter(byteRateLimiterConfig)
		}
		var messageRateLimiter *Tools.TokenBucketRateLimiter
		if messageRateLimiterConfig != nil {
			messageRateLimiter = Tools.NewTokenBucketRateLimiter(messageRateLimiterConfig)
		}

		if initializerFunc != nil {
			initializerFunc(websocketServer, websocketClient, identity, sessionId)
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

			object, err := deserializer(bytes)
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

			if err := validator(object); err != nil {
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

			return objectHandler(object, websocketServer, websocketClient, identity, sessionId)
		}
	}
}
