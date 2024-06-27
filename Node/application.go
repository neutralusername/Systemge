package Node

import (
	"Systemge/Message"
	"net/http"
)

type ApplicationConfig struct {
	ResolverAddress        string // *required*
	ResolverNameIndication string // *required*
	ResolverTLSCert        string // *required*

	HandleMessagesSequentially bool // default: false
}
type Application interface {
	GetAsyncMessageHandlers() map[string]AsyncMessageHandler
	GetSyncMessageHandlers() map[string]SyncMessageHandler
	GetCustomCommandHandlers() map[string]CustomCommandHandler
	OnStart(*Node) error
	OnStop(*Node) error
	GetApplicationConfig() ApplicationConfig
}
type AsyncMessageHandler func(*Node, *Message.Message) error
type SyncMessageHandler func(*Node, *Message.Message) (string, error)
type CustomCommandHandler func(*Node, []string) error

type HTTPComponentConfig struct {
	Port        string // *required*
	TlsCertPath string // *optional*
	TlsKeyPath  string // *optional*
}
type HTTPComponent interface {
	GetHTTPRequestHandlers() map[string]HTTPRequestHandler
	GetHTTPComponentConfig() HTTPComponentConfig
}
type HTTPRequestHandler func(w http.ResponseWriter, r *http.Request)

type WebsocketComponentConfig struct {
	Pattern     string // *required*
	Port        string // *required*
	TlsCertPath string // *optional*
	TlsKeyPath  string // *optional*
}
type WebsocketComponent interface {
	GetWebsocketMessageHandlers() map[string]WebsocketMessageHandler
	OnConnectHandler(*Node, *WebsocketClient)
	OnDisconnectHandler(*Node, *WebsocketClient)
	GetWebsocketComponentConfig() WebsocketComponentConfig
}
type WebsocketMessageHandler func(*Node, *WebsocketClient, *Message.Message) error

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
