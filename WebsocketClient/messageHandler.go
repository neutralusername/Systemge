package WebsocketClient

import "github.com/neutralusername/Systemge/Message"

type WebsocketMessageHandler func(*WebsocketClient, *Message.Message) error

type WebsocketMessageHandlers map[string]WebsocketMessageHandler
