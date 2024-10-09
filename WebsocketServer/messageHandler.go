package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func NewWebsocketTopicManager(config *Config.TopicManager, websocketMessageHandlers map[string]ObjectHandler, unknownWebsocketMessageHandler ObjectHandler) *Tools.TopicManager {
	topicHandlers := make(Tools.TopicHandlers)
	for topic, handler := range websocketMessageHandlers {
		topicHandlers[topic] = func(args ...any) (any, error) {
			message := args[0].(*Message.Message)
			websocketServer := args[1].(*WebsocketServer)
			websocketClient := args[2].(*WebsocketClient.WebsocketClient)
			identity := args[3].(string)
			sessionId := args[4].(string)
			return nil, handler(message, websocketServer, websocketClient, identity, sessionId)
		}
	}
	unknownTopicHandler := func(args ...any) (any, error) {
		message := args[0].(*Message.Message)
		websocketServer := args[1].(*WebsocketServer)
		websocketClient := args[2].(*WebsocketClient.WebsocketClient)
		identity := args[3].(string)
		sessionId := args[4].(string)
		return nil, unknownWebsocketMessageHandler(message, websocketServer, websocketClient, identity, sessionId)
	}
	return Tools.NewTopicManager(config, topicHandlers, unknownTopicHandler)
}
