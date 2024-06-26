package Client

import (
	"Systemge/Message"
	"net/http"
)

type Application interface {
	GetAsyncMessageHandlers() map[string]AsyncMessageHandler
	GetSyncMessageHandlers() map[string]SyncMessageHandler
	GetCustomCommandHandlers() map[string]CustomCommandHandler
	OnStart(*Client) error
	OnStop(*Client) error
}
type AsyncMessageHandler func(*Client, *Message.Message) error
type SyncMessageHandler func(*Client, *Message.Message) (string, error)
type CustomCommandHandler func(*Client, []string) error

type HTTPComponent interface {
	GetHTTPRequestHandlers() map[string]HTTPRequestHandler
}
type HTTPRequestHandler func(w http.ResponseWriter, r *http.Request)

type WebsocketComponent interface {
	GetWebsocketMessageHandlers() map[string]WebsocketMessageHandler
	OnConnectHandler(*Client, *WebsocketClient)
	OnDisconnectHandler(*Client, *WebsocketClient)
}
type WebsocketMessageHandler func(*Client, *WebsocketClient, *Message.Message) error

type WebsocketApplication interface {
	Application
	WebsocketComponent
}

type HTTPApplication interface {
	Application
	HTTPComponent
}

type WebsocketHTTPApplication interface {
	Application
	WebsocketComponent
	HTTPComponent
}
