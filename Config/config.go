package Config

import (
	"Systemge/TcpEndpoint"
	"Systemge/TcpServer"
	"Systemge/Utilities"
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
