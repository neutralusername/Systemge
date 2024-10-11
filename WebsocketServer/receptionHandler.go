package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/ReceptionHandler"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

type WebsocketServerReceptionHandlerFactory[T any] func(websocketServer *WebsocketServer[T], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) ReceptionHandler.ReceptionHandler
type WebsocketServerObjectHandler[T any] func(object T, websocketServer *WebsocketServer[T], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error
type WebsocketReceptionHandlerInitFunc[T any] func(websocketServer *WebsocketServer[T], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string)

func NewDefaultReceptionHandlerFactory[T any]() WebsocketServerReceptionHandlerFactory[T] {
	return func(websocketServer *WebsocketServer[T], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) ReceptionHandler.ReceptionHandler {
		return func(bytes []byte) error {
			return nil
		}
	}
}

func NewWebsocketTopicManager[T any](config *Config.TopicManager, topicObjectHandlers map[string]WebsocketServerObjectHandler[T], unknownObjectHandler WebsocketServerObjectHandler[T]) *Tools.TopicManager {
	topicHandlers := make(Tools.TopicHandlers)
	for topic, objectHandler := range topicObjectHandlers {
		topicHandlers[topic] = func(args ...any) (any, error) {
			object := args[0].(T)
			websocketServer := args[1].(*WebsocketServer[T])
			websocketClient := args[2].(*WebsocketClient.WebsocketClient)
			identity := args[3].(string)
			sessionId := args[4].(string)
			return nil, objectHandler(object, websocketServer, websocketClient, identity, sessionId)
		}
	}
	unknownTopicHandler := func(args ...any) (any, error) {
		message := args[0].(T)
		websocketServer := args[1].(*WebsocketServer[T])
		websocketClient := args[2].(*WebsocketClient.WebsocketClient)
		identity := args[3].(string)
		sessionId := args[4].(string)
		return nil, unknownObjectHandler(message, websocketServer, websocketClient, identity, sessionId)
	}
	return Tools.NewTopicManager(config, topicHandlers, unknownTopicHandler)
}

func NewValidationMessageReceptionHandlerFactory(byteRateLimiterConfig *Config.TokenBucketRateLimiter, messageRateLimiterConfig *Config.TokenBucketRateLimiter, messageValidatorConfig *Config.MessageValidator, topicManager *Tools.TopicManager, priorityQueue *Tools.PriorityTokenQueue[*Message.Message], topicPriorities map[string]uint32, topicTimeoutMs map[string]uint32) WebsocketServerReceptionHandlerFactory[*Message.Message] {
	objectDeserializer := func(messageBytes []byte) (*Message.Message, error) {
		return Message.Deserialize(messageBytes)
	}
	objectValidator := func(message *Message.Message) error {
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

	var messageHandler func(*Message.Message, *WebsocketServer[*Message.Message], *WebsocketClient.WebsocketClient, string, string) error
	var websocketReceptionHandlerInitFunc WebsocketReceptionHandlerInitFunc[*Message.Message]
	objectHandler := func(message *Message.Message, websocketServer *WebsocketServer[*Message.Message], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error {
		ReceptionHandler.NewChainObjecthandler(
			ReceptionHandler.NewValidationObjectHandler(objectValidator),
			ReceptionHandler.NewResponseObjectHandler(websocketServer.GetRequestResponseManager(), func(message *Message.Message) string {
				if message.IsResponse() {
					return message.GetSyncToken()
				}
				return ""
			}),
			ReceptionHandler.NewQueueObjectHandler(priorityQueue, func(message *Message.Message) (string, uint32, uint32) {
				priority := topicPriorities[message.GetTopic()]
				timeoutMs := topicTimeoutMs[message.GetTopic()]
				return "", priority, timeoutMs
			}),
		)
		if message.IsResponse() {
			// event
			websocketServer.GetRequestResponseManager().AddResponse(message.GetSyncToken(), message)
			return nil
		}
		return messageHandler(message, websocketServer, websocketClient, identity, sessionId)
	}

	return NewValidationReceptionHandlerFactory(byteRateLimiterConfig, messageRateLimiterConfig, objectDeserializer, objectHandler, websocketReceptionHandlerInitFunc)
}

func NewValidationReceptionHandlerFactory[T any](byteRateLimiterConfig *Config.TokenBucketRateLimiter, messageRateLimiterConfig *Config.TokenBucketRateLimiter, deserializer ReceptionHandler.ObjectDeserializer[T], websocketServerObjectHandler WebsocketServerObjectHandler[T], websocketReceptionHandlerInitFunc WebsocketReceptionHandlerInitFunc[T]) WebsocketServerReceptionHandlerFactory[T] {
	return func(websocketServer *WebsocketServer[T], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) ReceptionHandler.ReceptionHandler {
		objectHandler := func(object T) error {
			return websocketServerObjectHandler(object, websocketServer, websocketClient, identity, sessionId)
		}
		websocketReceptionHandlerInitFunc(websocketServer, websocketClient, identity, sessionId)
		var byteRateLimiter *Tools.TokenBucketRateLimiter
		if byteRateLimiterConfig != nil {
			byteRateLimiter = Tools.NewTokenBucketRateLimiter(byteRateLimiterConfig)
		}
		var messageRateLimiter *Tools.TokenBucketRateLimiter
		if messageRateLimiterConfig != nil {
			messageRateLimiter = Tools.NewTokenBucketRateLimiter(messageRateLimiterConfig)
		}

		return ReceptionHandler.NewReceptionHandler[T](
			ReceptionHandler.NewChainByteHandler(
				ReceptionHandler.ByteHandler[T](ReceptionHandler.NewMessageRateLimitByteHandler[T](messageRateLimiter)),
				ReceptionHandler.ByteHandler[T](ReceptionHandler.NewByteRateLimitByteHandler[T](byteRateLimiter)),
			),
			deserializer,
			objectHandler,
		)
	}
}
