package Config

import "github.com/gorilla/websocket"

type Node struct {
	Name string // *required*

	ErrorLogger   *Logger // *optional*
	WarningLogger *Logger // *optional*
	InfoLogger    *Logger // *optional*
	DebugLogger   *Logger // *optional*

	Mailer *Mailer // *optional*

	RandomizerSeed int64 // default: 0
}

type Systemge struct {
	HandleMessagesSequentially bool // default: false

	BrokerSubscribeDelayMs    uint64 // default: 0 (delay after failed broker subscription attempt)
	TopicResolutionLifetimeMs uint64 // default: 0
	SyncResponseTimeoutMs     uint64 // default: 0
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

	Blacklist []string // *optional*
	Whitelist []string // *optional* (if empty, all IPs are allowed)
}

type Http struct {
	Server *TcpServer // *required*

	Blacklist []string // *optional*
	Whitelist []string // *optional* (if empty, all IPs are allowed)
}
