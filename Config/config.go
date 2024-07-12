package Config

import (
	"Systemge/TcpEndpoint"
	"Systemge/TcpServer"
	"Systemge/Utilities"
)

type Node struct {
	Name             string            // *required*
	Logger           *Utilities.Logger // *required*
	ResolverEndpoint TcpEndpoint.TcpEndpoint

	BrokerSubscribeDelayMs    int // default: 0 (delay after failed broker subscription attempt)
	TopicResolutionLifetimeMs int // default: 0
	SyncResponseTimeoutMs     int // default: 0
	TcpTimeoutMs              int // default: 0 = block forever
}

type Application struct {
	HandleMessagesSequentially bool // default: false
}

type Websocket struct {
	Pattern                          string              // *required*
	Server                           TcpServer.TcpServer // *required*
	HandleClientMessagesSequentially bool                // default: false

	ClientMessageCooldownMs int // default: 0
	ClientWatchdogTimeoutMs int // default: 0
}

type HTTP struct {
	Server TcpServer.TcpServer // *required*
}

type Broker struct {
	Name                   string                  // *required*
	Logger                 *Utilities.Logger       // *required*
	ResolverConfigEndpoint TcpEndpoint.TcpEndpoint // *required*

	SyncResponseTimeoutMs int // default: 0
	TcpTimeoutMs          int // default: 0 = block forever

	Server   TcpServer.TcpServer     // *required*
	Endpoint TcpEndpoint.TcpEndpoint // *required*

	ConfigServer TcpServer.TcpServer // *required*

	SyncTopics  []string
	AsyncTopics []string
}

type Resolver struct {
	Name   string // *required*
	Logger *Utilities.Logger

	TcpTimeoutMs int // default: 0 = block forever

	Server       TcpServer.TcpServer // *required*
	ConfigServer TcpServer.TcpServer // *required*
}

type Spawner struct {
	IsSpawnedNodeTopicSync bool              // default: false
	SpawnedNodeLogger      *Utilities.Logger // *required*

	ResolverEndpoint     TcpEndpoint.TcpEndpoint // *required*
	BrokerConfigEndpoint TcpEndpoint.TcpEndpoint // *required*
}
