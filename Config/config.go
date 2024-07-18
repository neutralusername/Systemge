package Config

import (
	"Systemge/Tcp"

	"golang.org/x/oauth2"
)

type Node struct {
	Name   string // *required*
	Logger Logger // *optional*
	Mailer Mailer // *optional*
}

type Systemge struct {
	HandleMessagesSequentially bool // default: false

	BrokerSubscribeDelayMs    uint64 // default: 0 (delay after failed broker subscription attempt)
	TopicResolutionLifetimeMs uint64 // default: 0
	SyncResponseTimeoutMs     uint64 // default: 0
	TcpTimeoutMs              uint64 // default: 0 = block forever

	ResolverEndpoint Tcp.Endpoint // *required*
}

type Websocket struct {
	Pattern                          string     // *required*
	Server                           Tcp.Server // *required*
	HandleClientMessagesSequentially bool       // default: false

	ClientMessageCooldownMs uint64 // default: 0
	ClientWatchdogTimeoutMs uint64 // default: 0
}

type HTTP struct {
	Server Tcp.Server // *required*
}

type Logger struct {
	InfoPath    string // *required*
	WarningPath string // *required*
	ErrorPath   string // *required*
	DebugPath   string // *required*
	QueueBuffer int    // default: 0
}

type Mailer struct {
	SmtpHost       string   // *required*
	SmtpPort       uint16   // *required*
	SenderEmail    string   // *required*
	SenderPassword string   // *required*
	Recipients     []string // *required*
	Logger         Logger   // *required*
}

type Oauth2 struct {
	Server                     Tcp.Server                                                                  // *required*
	AuthPath                   string                                                                      // *required*
	AuthCallbackPath           string                                                                      // *required*
	OAuth2Config               *oauth2.Config                                                              // *required*
	AuthRedirectUrl            string                                                                      // *optional*
	CallbackSuccessRedirectUrl string                                                                      // *required*
	CallbackFailureRedirectUrl string                                                                      // *required*
	TokenHandler               func(*oauth2.Config, *oauth2.Token) (string, map[string]interface{}, error) // *required
	SessionLifetimeMs          uint64                                                                      // default: 0
	RandomizerSeed             int64                                                                       // *required*
	Oauth2State                string                                                                      // *required*
}

type Broker struct {
	Server       Tcp.Server   // *required*
	Endpoint     Tcp.Endpoint // *required*
	ConfigServer Tcp.Server   // *required*

	SyncTopics  []string
	AsyncTopics []string

	ResolverConfigEndpoint Tcp.Endpoint // *required*

	SyncResponseTimeoutMs uint64 // default: 0
	TcpTimeoutMs          uint64 // default: 0 = block forever

	MaxMessageSize uint64 // default: 0 = unlimited
	MaxOriginSize  int    // default: 0 = unlimited
	MaxPayloadSize int    // default: 0 = unlimited
	MaxTopicSize   int    // default: 0 = unlimited
	MaxSyncKeySize int    // default: 0 = unlimited
}

type Resolver struct {
	Server       Tcp.Server // *required*
	ConfigServer Tcp.Server // *required*

	TcpTimeoutMs uint64 // default: 0 = block forever

	MaxMessageSize uint64 // default: 0 = unlimited
	MaxPayloadSize int    // default: 0 = unlimited
	MaxOriginSize  int    // default: 0 = unlimited
	MaxTopicSize   int    // default: 0 = unlimited
}

type Spawner struct {
	IsSpawnedNodeTopicSync bool   // default: false
	Logger                 Logger // *required*

	ResolverEndpoint     Tcp.Endpoint // *required*
	BrokerConfigEndpoint Tcp.Endpoint // *required*
}
