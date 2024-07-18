package Config

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

	ResolverEndpoint TcpEndpoint // *required*
}

type Websocket struct {
	Pattern                          string    // *required*
	Server                           TcpServer // *required*
	HandleClientMessagesSequentially bool      // default: false

	ClientMessageCooldownMs uint64 // default: 0
	ClientWatchdogTimeoutMs uint64 // default: 0
}

type HTTP struct {
	Server TcpServer // *required*
}
