package Config

import "golang.org/x/oauth2"

type Oauth2 struct {
	Server                     *TcpServer                                                                  // *required*
	AuthPath                   string                                                                      // *required*
	AuthCallbackPath           string                                                                      // *required*
	OAuth2Config               *oauth2.Config                                                              // *required*
	AuthRedirectUrl            string                                                                      // *optional*
	CallbackSuccessRedirectUrl string                                                                      // *required*
	CallbackFailureRedirectUrl string                                                                      // *required*
	TokenHandler               func(*oauth2.Config, *oauth2.Token) (string, map[string]interface{}, error) // *required
	SessionLifetimeMs          uint64                                                                      // default: 0
	Oauth2State                string                                                                      // *required*
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

type Spawner struct {
	IsSpawnedNodeTopicSync bool    // default: false
	ErrorLogger            *Logger // *required*
	WarningLogger          *Logger // *required*
	InfoLogger             *Logger // *required*
	DebugLogger            *Logger // *required*
	Mailer                 *Mailer // *required*

	ResolverEndpoint     *TcpEndpoint // *required*
	BrokerConfigEndpoint *TcpEndpoint // *required*
}

// Server applies to both http and websocket besides the fact that websocket is hardcoded to port 18251
type Dashboard struct {
	Server                         *TcpServer // *required*
	AutoStart                      bool       // default: false
	HeapUpdateIntervalMs           uint64     // default: 0 = disabled
	NodeStatusIntervalMs           uint64     // default: 0 = disabled
	NodeSystemgeCountersIntervalMs uint64     // default: 0 = disabled
	NodeWebsocketCounterIntervalMs uint64     // default: 0 = disabled
}
