package Config

import (
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

type Node struct {
	Name string // *required*

	Mailer        *Tools.Mailer // *optional*
	InfoLogger    *Tools.Logger // *optional*
	WarningLogger *Tools.Logger // *optional*
	ErrorLogger   *Tools.Logger // *optional*
	DebugLogger   *Tools.Logger // *optional*

	RandomizerSeed int64 // default: 0
}

type Systemge struct {
	HandleMessagesSequentially bool // default: false

	BrokerSubscribeDelayMs    uint64 // default: 0 (delay after failed broker subscription attempt)
	TopicResolutionLifetimeMs uint64 // default: 0 = resolve topic every time a message is sent
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
