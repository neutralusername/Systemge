package WebsocketServer

//type WebsocketServerObjectHandler[T any] func(object T, websocketServer *WebsocketServer[T], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error
//type WebsocketReceptionHandlerInitFunc[T any] func(websocketServer *WebsocketServer[T], websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string)

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
