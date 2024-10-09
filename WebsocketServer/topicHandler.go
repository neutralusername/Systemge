package WebsocketServer

/* func NewWebsocketTopicManager(config *Config.TopicManager, websocketTopicHandlers map[string]WebsocketTopicHandler, unknownWebsocketTopicHandler WebsocketTopicHandler) *Tools.TopicManager {
	topicHandlers := make(Tools.TopicHandlers)
	for topic, handler := range websocketTopicHandlers {
		topicHandlers[topic] = func(args ...any) (any, error) {
			message := args[0].(*Message.Message)
			websocketServer := args[1].(*WebsocketServer)
			websocketClient := args[2].(*WebsocketClient.WebsocketClient)
			identity := args[3].(string)
			sessionId := args[4].(string)
			priority := args[5].(uint32)
			timeoutMs := args[6].(uint64)
			return nil, handler(message, websocketServer, websocketClient, identity, sessionId, priority, timeoutMs)
		}
	}
	unknownTopicHandler := func(args ...any) (any, error) {
		message := args[0].(*Message.Message)
		websocketServer := args[1].(*WebsocketServer)
		websocketClient := args[2].(*WebsocketClient.WebsocketClient)
		identity := args[3].(string)
		sessionId := args[4].(string)
		priority := args[5].(uint32)
		timeoutMs := args[6].(uint64)
		return nil, unknownWebsocketTopicHandler(message, websocketServer, websocketClient, identity, sessionId, priority, timeoutMs)
	}
	return Tools.NewTopicManager(config, topicHandlers, unknownTopicHandler)
}
*/
