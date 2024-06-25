package Client

import (
	"Systemge/Message"
	"Systemge/WebsocketClient"
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

type HTTPApplication interface {
	GetHTTPRequestHandlers() map[string]HTTPRequestHandler
}
type HTTPRequestHandler func(w http.ResponseWriter, r *http.Request)

type WebsocketApplication interface {
	GetWebsocketMessageHandlers() map[string]WebsocketMessageHandler
	OnConnectHandler(*Client, *WebsocketClient.Client)
	OnDisconnectHandler(*Client, *WebsocketClient.Client)
}
type WebsocketMessageHandler func(*Client, *WebsocketClient.Client, *Message.Message) error

type CompositeApplicationWebsocket interface {
	Application
	WebsocketApplication
}

type CompositeApplicationHTTP interface {
	Application
	HTTPApplication
}

type CompositeApplicationWebsocketHTTP interface {
	Application
	WebsocketApplication
	HTTPApplication
}
