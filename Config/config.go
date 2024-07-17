package Config

import (
	"Systemge/TcpEndpoint"
	"Systemge/TcpServer"
	"Systemge/Utilities"

	"golang.org/x/oauth2"
)

type Node struct {
	Name   string            // *required*
	Logger *Utilities.Logger // *required*
}

type Systemge struct {
	HandleMessagesSequentially bool // default: false

	BrokerSubscribeDelayMs    uint64 // default: 0 (delay after failed broker subscription attempt)
	TopicResolutionLifetimeMs uint64 // default: 0
	SyncResponseTimeoutMs     uint64 // default: 0
	TcpTimeoutMs              uint64 // default: 0 = block forever

	ResolverEndpoint TcpEndpoint.TcpEndpoint // *required*
}

type Websocket struct {
	Pattern                          string              // *required*
	Server                           TcpServer.TcpServer // *required*
	HandleClientMessagesSequentially bool                // default: false

	ClientMessageCooldownMs uint64 // default: 0
	ClientWatchdogTimeoutMs uint64 // default: 0
}

type HTTP struct {
	Server TcpServer.TcpServer // *required*
}

type Oauth2 struct {
	Name                    string
	Port                    uint16
	AuthPath                string
	AuthCallbackPath        string
	OAuth2Config            *oauth2.Config
	SucessCallbackRedirect  string
	FailureCallbackRedirect string
	Logger                  *Utilities.Logger
	TokenHandler            func(*oauth2.Config, *oauth2.Token) (string, map[string]interface{}, error)
	SessionLifetimeMs       uint64
	Randomizer              *Utilities.Randomizer
	Oauth2State             string
}

type Broker struct {
	Server       TcpServer.TcpServer     // *required*
	Endpoint     TcpEndpoint.TcpEndpoint // *required*
	ConfigServer TcpServer.TcpServer     // *required*

	SyncTopics  []string
	AsyncTopics []string

	ResolverConfigEndpoint TcpEndpoint.TcpEndpoint // *required*

	SyncResponseTimeoutMs uint64 // default: 0
	TcpTimeoutMs          uint64 // default: 0 = block forever

	MaxMessageSize uint64 // default: 0 = unlimited
	MaxOriginSize  int    // default: 0 = unlimited
	MaxPayloadSize int    // default: 0 = unlimited
	MaxTopicSize   int    // default: 0 = unlimited
	MaxSyncKeySize int    // default: 0 = unlimited
}

type Resolver struct {
	Server       TcpServer.TcpServer // *required*
	ConfigServer TcpServer.TcpServer // *required*

	TcpTimeoutMs uint64 // default: 0 = block forever

	MaxMessageSize uint64 // default: 0 = unlimited
	MaxPayloadSize int    // default: 0 = unlimited
	MaxOriginSize  int    // default: 0 = unlimited
	MaxTopicSize   int    // default: 0 = unlimited
}

type Spawner struct {
	IsSpawnedNodeTopicSync bool              // default: false
	SpawnedNodeLogger      *Utilities.Logger // *required*

	ResolverEndpoint     TcpEndpoint.TcpEndpoint // *required*
	BrokerConfigEndpoint TcpEndpoint.TcpEndpoint // *required*
}
