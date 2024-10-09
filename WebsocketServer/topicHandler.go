package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

type WebsocketTopicHandler func(message *Message.Message, websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error

func NewWebsocketTopicManager(config *Config.TopicManager, topicHandlers map[string]WebsocketTopicHandler, unknownTopicHandler WebsocketTopicHandler) *Tools.TopicManager {

	return Tools.NewTopicManager(config, convertedTopicHandlers, convertedUnknownTopicHandler)
}
