package Config

import (
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

type Node struct {
	Name string // *required*

	Mailer                *Tools.Mailer // *optional*
	InfoLogger            *Tools.Logger // *optional*
	InternalInfoLogger    *Tools.Logger // *optional*
	WarningLogger         *Tools.Logger // *optional*
	InternalWarningLogger *Tools.Logger // *optional*
	ErrorLogger           *Tools.Logger // *optional*
	DebugLogger           *Tools.Logger // *optional*

	RandomizerSeed int64 // default: 0
}

type Systemge struct {
	HandleMessagesSequentially bool // default: false

	BrokerSubscribeDelayMs    uint64 // default: 0 (delay after failed broker subscription attempt)
	TopicResolutionLifetimeMs int64  // default: <0 = infinite; 0 = resolve every time; >0 = resolve every n ms
	SyncResponseTimeoutMs     uint64 // default: 0 = sync messages are not supported
	TcpTimeoutMs              uint64 // default: 0 = block forever
	MaxSubscribeAttempts      uint64 // default: 0 = infinite

	ResolverEndpoint *TcpEndpoint // *required*
}

type Websocket struct {
	Pattern string     // *required*
	Server  *TcpServer // *required*

	HandleClientMessagesSequentially bool   // default: false
	ClientMessageCooldownMs          uint64 // default: 0
	ClientWatchdogTimeoutMs          uint64 // default: 0

	Upgrader *websocket.Upgrader // *required*

}

type HTTP struct {
	Server *TcpServer // *required*
}

type Broker struct {
	Server       *TcpServer   // *required*
	Endpoint     *TcpEndpoint // *required*
	ConfigServer *TcpServer   // *required*

	SyncTopics  []string
	AsyncTopics []string

	ResolverConfigEndpoint *TcpEndpoint // *required*

	SyncResponseTimeoutMs uint64 // default: 0
	TcpTimeoutMs          uint64 // default: 0 = block forever

	IncomingMessageByteLimit uint64 // default: 0 = unlimited
	MaxOriginSize            int    // default: 0 = unlimited
	MaxPayloadSize           int    // default: 0 = unlimited
	MaxTopicSize             int    // default: 0 = unlimited
	MaxSyncKeySize           int    // default: 0 = unlimited
}

type Resolver struct {
	Server       *TcpServer // *required*
	ConfigServer *TcpServer // *required*

	TcpTimeoutMs uint64 // default: 0 = block forever

	IncomingMessageByteLimit uint64 // default: 0 = unlimited
	MaxPayloadSize           int    // default: 0 = unlimited
	MaxOriginSize            int    // default: 0 = unlimited
	MaxTopicSize             int    // default: 0 = unlimited
}
