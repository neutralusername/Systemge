package Config

import (
	"Systemge/TcpEndpoint"
	"Systemge/TcpServer"
)

type Node struct {
	Name             string // *required*
	LoggerPath       string // *required*
	ResolverEndpoint TcpEndpoint.TcpEndpoint

	HeartbeatIntervalMs   int // default: 0
	SyncResponseTimeoutMs int // default: 0
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
	LoggerPath             string                  // *required*
	ResolverConfigEndpoint TcpEndpoint.TcpEndpoint // *required*

	DeliverImmediately   bool // default: false
	ConnectionTimeoutMs  int  // default: 0
	SyncRequestTimeoutMs int  // default: 0

	Server   TcpServer.TcpServer     // *required*
	Endpoint TcpEndpoint.TcpEndpoint // *required*

	ConfigServer TcpServer.TcpServer // *required*

	SyncTopics  []string
	AsyncTopics []string
}

type Resolver struct {
	Name       string // *required*
	LoggerPath string // *required*

	Server   TcpServer.TcpServer     // *required*
	Endpoint TcpEndpoint.TcpEndpoint // *required*

	ConfigServer TcpServer.TcpServer // *required*
}

type Spawner struct {
	IsSpawnedNodeTopicSync bool   // default: false
	SpawnedNodeLoggerPath  string // *required*

	ResolverEndpoint     TcpEndpoint.TcpEndpoint // *required*
	BrokerConfigEndpoint TcpEndpoint.TcpEndpoint // *required*
}
