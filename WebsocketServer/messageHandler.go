package WebsocketServer

import "github.com/neutralusername/Systemge/Message"

type WebsocketMessageHandler func(*WebsocketConnection, *Message.Message) error

type WebsocketMessageHandlers map[string]WebsocketMessageHandler
