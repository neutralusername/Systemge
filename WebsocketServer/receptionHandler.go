package WebsocketServer

import (
	"errors"

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

func NewValidationMessageReceptionHandlerFactory(byteRateLimiterConfig *Config.TokenBucketRateLimiter, messageRateLimiterConfig *Config.TokenBucketRateLimiter, messageValidatorConfig *Config.MessageValidator, topicManager *Tools.TopicManager, priorityQueue *Tools.PriorityTokenQueue[*Message.Message]) ReceptionHandlerFactory {
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

	objectHandler := func(object any, websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error {
		message := object.(*Message.Message)

		// queue(-config) and topic-priority&timeout missing
		// feed message into queue, if queue enabled, or directly into topic manager

		if noQueue {
			response, err := topicManager.HandleTopic(message.GetTopic(), message, websocketServer, websocketClient, identity, sessionId)
			if err != nil {
				// event
				return err
			}
		} else {
			priorityQueue.Push("", message, 0, 0)
		}

		// handle response

		return nil
	}
	return NewValidationReceptionHandlerFactory(byteRateLimiterConfig, messageRateLimiterConfig, objectDeserializer, objectValidator, objectHandler)
}

func NewValidationReceptionHandlerFactory(byteRateLimiterConfig *Config.TokenBucketRateLimiter, messageRateLimiterConfig *Config.TokenBucketRateLimiter, deserializer ObjectDeserializer, validator ObjectValidator, objectHandler ObjectHandler) ReceptionHandlerFactory {
	return func(websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) ReceptionHandler {
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

/*

func (connection *WebsocketClient) handleMessageReception(messageBytes []byte, behaviour string) (*Message.Message, error) {

	if event := connection.onEvent(Event.NewInfo(
		Event.SendingToChannel,
		"sending message to channel",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.HandleMessageReception,
			Event.Behaviour:    behaviour,
			Event.ChannelType:  Event.MessageChannel,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	)); !event.IsInfo() {
		connection.rejectedMessagesReceived.Add(1)
		connection.messageChannelSemaphore.ReleaseBlocking()
		return event.GetError()
	}
	connection.priorityTokenQueue.AddItem("", message, 0, 0)

	connection.onEvent(Event.NewInfoNoOption(
		Event.HandledMessageReception,
		"handled message reception",
		Event.Context{
			Event.Circumstance: Event.HandleMessageReception,
			Event.Behaviour:    behaviour,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	))

	return nil
}

func (connection *WebsocketClient) validateMessage(message *Message.Message) error {
	if len(message.GetSyncToken()) != 0 {
		return errors.New("message contains sync token")
	}
	if len(message.GetTopic()) == 0 {
		return errors.New("message missing topic")
	}
	if maxTopicSize := connection.config.MaxTopicSize; maxTopicSize > 0 && len(message.GetTopic()) > maxTopicSize {
		return errors.New("message topic exceeds maximum size")
	}
	if maxPayloadSize := connection.config.MaxPayloadSize; maxPayloadSize > 0 && len(message.GetPayload()) > maxPayloadSize {
		return errors.New("message payload exceeds maximum size")
	}
	return nil
}
*/

/* func (connection *WebsocketClient) RetrieveNextMessage(waitDurationMs uint64) (*Message.Message, error) {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()

	if connection.messageHandlingLoopStopChannel != nil {
		connection.onEvent(Event.NewWarningNoOption(
			Event.MessageHandlingLoopAlreadyStarted,
			"message handling loop is registered",
			Event.Context{
				Event.Circumstance: Event.RetrieveNextMessage,
				Event.ChannelType:  Event.MessageChannel,
			},
		))
		return nil, errors.New("message handling loop is registered")
	}

	if event := connection.onEvent(Event.NewInfo(
		Event.ReceivingFromChannel,
		"receiving from message channel",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.RetrieveNextMessage,
			Event.ChannelType:  Event.MessageChannel,
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	var timeout <-chan time.Time
	if waitDurationMs > 0 {
		timeout = time.After(time.Duration(waitDurationMs) * time.Millisecond)
	}
	select {
	case message := <-connection.messageChannel:
		if message == nil {
			connection.onEvent(Event.NewWarningNoOption(
				Event.ReceivedNilValueFromChannel,
				"received nil value from message channel",
				Event.Context{
					Event.Circumstance: Event.RetrieveNextMessage,
					Event.ChannelType:  Event.MessageChannel,
				},
			))
			return nil, errors.New("received nil value from message channel")
		}
		if event := connection.onEvent(Event.NewInfo(
			Event.ReceivedFromChannel,
			"received message from message channel",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.RetrieveNextMessage,
				Event.ChannelType:  Event.MessageChannel,
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
		)); !event.IsInfo() {
			return nil, event.GetError()
		}
		connection.messageChannelSemaphore.ReleaseBlocking()
		return message, nil
	case <-timeout:
		connection.onEvent(Event.NewWarningNoOption(
			Event.Timeout,
			"timeout while waiting for message",
			Event.Context{
				Event.Circumstance: Event.RetrieveNextMessage,
				Event.ChannelType:  Event.MessageChannel,
			},
		))
		return nil, errors.New("timeout while waiting for message")
	}
}


func (server *WebsocketServer) toTopicHandler(handler WebsocketClient.WebsocketMessageHandler) Tools.TopicHandler {
	return func(args ...any) (any, error) {
		websocketConnection := args[0].(*WebsocketClient.WebsocketClient)
		message := args[1].(*Message.Message)

		if server.eventHandler != nil {
			if event := server.onEvent(Event.NewInfo(
				Event.HandlingMessage,
				"handling websocketConnection message",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance: Event.HandleMessage,
					Event.Identity:     websocketConnection.GetName(),
					Event.Address:      websocketConnection.GetAddress(),
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
			)); !event.IsInfo() {
				return nil, errors.New("event cancelled")
			}
		}

		err := handler(websocketConnection, message)
		if err != nil {
			if server.eventHandler != nil {
				server.onEvent(Event.NewWarningNoOption(
					Event.HandleMessageFailed,
					err.Error(),
					Event.Context{
						Event.Circumstance: Event.HandleMessage,
						Event.Identity:     websocketConnection.GetName(),
						Event.Address:      websocketConnection.GetAddress(),
						Event.Topic:        message.GetTopic(),
						Event.Payload:      message.GetPayload(),
						Event.SyncToken:    message.GetSyncToken(),
					},
				))
			}
			return nil, err
		}

		if server.eventHandler != nil {
			server.onEvent(Event.NewInfoNoOption(
				Event.HandledMessage,
				"handled websocketConnection message",
				Event.Context{
					Event.Circumstance: Event.HandleMessage,
					Event.Identity:     websocketConnection.GetName(),
					Event.Address:      websocketConnection.GetAddress(),
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
			))
		}

		return nil, nil
	}
}

		topicHandlers := make(Tools.TopicHandlers)
	   	for topic, handler := range messageHandlers {
	   		topicHandlers[topic] = server.toTopicHandler(handler)
	   	}
	   	server.topicManager = Tools.NewTopicManager(config.TopicManagerConfig, topicHandlers, nil)

// A started loop will run until stopChannel receives a value (or is closed) or connection.GetNextMessage returns an error.
// errorChannel will send all errors that occur during message processing.
func (connection *WebsocketClient) StartMessageHandlingLoop(topicManager *Tools.TopicManager, sequentially bool) error {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()

	var behaviour string
	if sequentially {
		behaviour = Event.Sequential
	} else {
		behaviour = Event.Concurrent
	}

	if event := connection.onEvent(Event.NewInfo(
		Event.MessageHandlingLoopStarting,
		"starting message handling loop",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.MessageHandlingLoopStart,
			Event.Behaviour:    behaviour,
			Event.ChannelType:  Event.MessageChannel,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if connection.messageHandlingLoopStopChannel != nil {
		connection.onEvent(Event.NewWarningNoOption(
			Event.MessageHandlingLoopAlreadyStarted,
			"message handling loop already started",
			Event.Context{
				Event.Circumstance: Event.MessageHandlingLoopStart,
				Event.Behaviour:    behaviour,
				Event.ChannelType:  Event.MessageChannel,
			},
		))
		return errors.New("message handling loop already started")
	}
	stopChannel := make(chan bool)
	connection.messageHandlingLoopStopChannel = stopChannel

	go connection.messageHandlingLoop(stopChannel, topicManager, sequentially, behaviour)

	if event := connection.onEvent(Event.NewInfo(
		Event.MessageHandlingLoopStarted,
		"message handling loop started",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.MessageHandlingLoopStart,
			Event.Behaviour:    behaviour,
			Event.ChannelType:  Event.MessageChannel,
		},
	)); !event.IsInfo() {
		close(connection.messageHandlingLoopStopChannel)
		connection.messageHandlingLoopStopChannel = nil
		return event.GetError()
	}

	return nil
}

func (connection *WebsocketClient) messageHandlingLoop(stopChannel chan bool, topicManager *Tools.TopicManager, sequentially bool, behaviour string) {
	if event := connection.onEvent(Event.NewInfo(
		Event.ReceivingFromChannel,
		"message handling loop running",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.MessageHandlingLoop,
			Event.Behaviour:    behaviour,
			Event.ChannelType:  Event.MessageChannel,
		},
	)); !event.IsInfo() {
		connection.StopMessageHandlingLoop()
		return
	}

	for {
		select {
		case <-stopChannel:
			return
		case message := <-connection.messageChannel:

			if message == nil {
				connection.onEvent(Event.NewInfoNoOption(
					Event.ReceivedNilValueFromChannel,
					"received nil value from message channel",
					Event.Context{
						Event.Circumstance: Event.MessageHandlingLoop,
						Event.ChannelType:  Event.MessageChannel,
						Event.Behaviour:    behaviour,
					},
				))
				connection.StopMessageHandlingLoop()
				return
			}

			connection.messageChannelSemaphore.ReleaseBlocking()
			if event := connection.onEvent(Event.NewInfo(
				Event.ReceivedFromChannel,
				"received message from message channel",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance: Event.MessageHandlingLoop,
					Event.Behaviour:    behaviour,
					Event.ChannelType:  Event.MessageChannel,
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
			)); !event.IsInfo() {
				connection.StopMessageHandlingLoop()
				return
			}

			if sequentially {
				topicManager.HandleTopic(message.GetTopic(), message, connection)
			} else {
				go func() {
					topicManager.HandleTopic(message.GetTopic(), message, connection)
				}()
			}
		}
	}
}

func (connection *WebsocketClient) StopMessageHandlingLoop() error {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()

	if event := connection.onEvent(Event.NewInfo(
		Event.MessageHandlingLoopStopping,
		"stopping message handling loop",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.MessageHandlingLoopStop,
			Event.ChannelType:  Event.MessageChannel,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if connection.messageHandlingLoopStopChannel == nil {
		connection.onEvent(Event.NewWarningNoOption(
			Event.MessageHandlingLoopAlreadyStopped,
			"message handling loop already stopped",
			Event.Context{
				Event.Circumstance: Event.MessageHandlingLoopStop,
				Event.ChannelType:  Event.MessageChannel,
			},
		))
		return errors.New("message handling loop already stopped")
	}

	close(connection.messageHandlingLoopStopChannel)
	connection.messageHandlingLoopStopChannel = nil

	connection.onEvent(Event.NewInfoNoOption(
		Event.MessageHandlingLoopStopped,
		"message handling loop stopped",
		Event.Context{
			Event.Circumstance: Event.MessageHandlingLoopStop,
			Event.ChannelType:  Event.MessageChannel,
		},
	))
	return nil
}

func (connection *WebsocketClient) IsMessageHandlingLoopStarted() bool {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()
	return connection.messageHandlingLoopStopChannel != nil
}

func (connection *WebsocketClient) AvailableMessageCount() uint32 {
	return connection.messageChannelSemaphore.AvailableAcquires()
}
*/
