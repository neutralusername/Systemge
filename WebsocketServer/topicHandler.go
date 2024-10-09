package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

type TopicHandler func(message *Message.Message, websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error

func NewWebsocketTopicManager(config *Config.TopicManager, topicHandlers map[string]TopicHandler, unknownTopicHandler TopicHandler) *Tools.TopicManager {

	return Tools.NewTopicManager(config, convertedTopicHandlers, convertedUnknownTopicHandler)
}
