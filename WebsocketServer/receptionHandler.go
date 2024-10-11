package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

//type WebsocketServerObjectHandler[T any] func(object T, websocketServer *WebsocketServer[T], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error
//type WebsocketReceptionHandlerInitFunc[T any] func(websocketServer *WebsocketServer[T], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string)

func NewDefaultReceptionHandlerFactory[T any]() *Tools.ReceptionHandlerFactory[*structName123] {
	return func(
	/* 	websocketServer *WebsocketServer[T],
	websocketClient *WebsocketClient.WebsocketClient,
	identity string,
	sessionId string, */
	) Tools.ReceptionHandler {

		return func(bytes []byte) error {
			return nil
		}
	}
}

/* func NewWebsocketTopicManager[P any, R any](
	config *Config.TopicManager,
	topicObjectHandlers map[string]WebsocketServerObjectHandler[P],
	unknownObjectHandler WebsocketServerObjectHandler[P],
) *Tools.TopicManager[P, R] {

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
} */

func NewValidationMessageReceptionHandlerFactory[S any](
	byteRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageValidatorConfig *Config.MessageValidator,
	//topicManager *Tools.TopicManager,
	priorityQueue *Tools.PriorityTokenQueue[*Message.Message],
	obtainEnqueueConfigs Tools.ObtainEnqueueConfigs[*Message.Message],
	requestResponseManager *Tools.RequestResponseManager[*Message.Message],
) Tools.ReceptionHandlerFactory[S] {

	byteHandlers := []Tools.ByteHandler[*Message.Message]{}
	if byteRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, Tools.NewByteRateLimitByteHandler[*Message.Message](Tools.NewTokenBucketRateLimiter(byteRateLimiterConfig)))
	}
	if messageRateLimiterConfig != nil {
		byteHandlers = append(byteHandlers, Tools.NewMessageRateLimitByteHandler[*Message.Message](Tools.NewTokenBucketRateLimiter(messageRateLimiterConfig)))
	}

	objectDeserializer := func(messageBytes []byte) (*Message.Message, error) {
		return Message.Deserialize(messageBytes)
	}

	objectHandlers := []Tools.ObjectHandler[*Message.Message]{}
	if messageValidatorConfig != nil {
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
		objectHandlers = append(objectHandlers, Tools.NewValidationObjectHandler(objectValidator))
	}
	if requestResponseManager != nil {
		obtainResponseToken := func(message *Message.Message) string {
			if message.IsResponse() {
				return message.GetSyncToken()
			}
			return ""
		}
		objectHandlers = append(objectHandlers, Tools.NewResponseObjectHandler(requestResponseManager, obtainResponseToken))
	}
	if priorityQueue != nil {
		if obtainEnqueueConfigs == nil {
			obtainEnqueueConfigs = func(message *Message.Message) (string, uint32, uint32) {
				return "", 0, 0
			}
		}
		objectHandlers = append(objectHandlers, Tools.NewQueueObjectHandler(priorityQueue, obtainEnqueueConfigs))

	}

	return NewValidationReceptionHandlerFactory[*Message.Message, S](
		Tools.NewChainByteHandler(byteHandlers...),
		objectDeserializer,
		Tools.NewChainObjecthandler(objectHandlers...),
	)
}

func NewValidationReceptionHandlerFactory[T any, S any](
	byteHandler Tools.ByteHandler[T],
	deserializer Tools.ObjectDeserializer[T],
	objectHandler Tools.ObjectHandler[T],
) Tools.ReceptionHandlerFactory[S] {

	return func(
		/* 	websocketServer *WebsocketServer[T],
		websocketClient *WebsocketClient.WebsocketClient,
		identity string,
		sessionId string, */
		structName123 S,
	) Tools.ReceptionHandler {

		return Tools.NewReceptionHandler[T](
			byteHandler,
			deserializer,
			objectHandler,
			structName123,
		)
	}
}

/*
var websocketReceptionHandlerInitFunc WebsocketReceptionHandlerInitFunc[*Message.Message]
		if topicManager != nil {
		handleTopic = func(message *Message.Message, websocketServer *WebsocketServer[*Message.Message], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error {
			// event

			response, err := topicManager.Handle(message.GetTopic(), message, websocketServer, websocketClient, identity, sessionId)
			if err != nil {
				// event
				return err
			}

			if response != nil {
				message, ok := response.(*Message.Message)
				if !ok {
					// event
					return errors.New("invalid response type")
				}
				if err := websocketClient.Write(message.Serialize(), websocketServer.config.WriteTimeoutMs); err != nil {
					// event
				}
			}
			// event
			return nil
		}
	}
	if priorityQueue != nil {
		if topicManager != nil {
			websocketReceptionHandlerInitFunc = func(websocketServer *WebsocketServer[*Message.Message], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) {
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
			}
		}
		mutex := &sync.Mutex{}
		messageHandler = func(message *Message.Message, websocketServer *WebsocketServer[*Message.Message], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error {
			mutex.Lock()
			priority := topicPriorities[message.GetTopic()]
			timeoutMs := topicTimeoutMs[message.GetTopic()]
			mutex.Unlock()
			// event
			return priorityQueue.Push("", message, priority, timeoutMs)
		}
	} else {
		if topicManager != nil {
			messageHandler = func(message *Message.Message, websocketServer *WebsocketServer[*Message.Message], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error {
				return handleTopic(message, websocketServer, websocketClient, identity, sessionId)
			}
		} else {
			messageHandler = func(message *Message.Message, websocketServer *WebsocketServer[*Message.Message], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error {
				return nil
			}
		}
	}


*/
