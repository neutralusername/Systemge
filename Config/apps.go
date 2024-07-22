package Config

import "golang.org/x/oauth2"

type Oauth2 struct {
	Http                       *Http                                                                       // *required*
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

	BrokerWhitelist []string // *optional* (if empty, all IPs are allowed)
	BrokerBlacklist []string // *optional*

	ConfigWhitelist []string // *optional* (if empty, all IPs are allowed)
	ConfigBlacklist []string // *optional*

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

	ResolverWhitelist []string // *optional* (if empty, all IPs are allowed)
	ResolverBlacklist []string // *optional*

	ConfigWhitelist []string // *optional* (if empty, all IPs are allowed)
	ConfigBlacklist []string // *optional*

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

// websocket and http share the same Http config (tls, blacklist, whitelist) besides the port, which only applies to http
// websocket runs on port 18251 which is hardcoded
type Dashboard struct {
	Http                   *Http  // default: nil (app)
	StatusUpdateIntervalMs uint64 // default: 0 = disabled
}
