package Config

import (
	"encoding/json"

	"golang.org/x/oauth2"
)

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

type Spawner struct {
	IsSpawnedNodeTopicSync      bool   // default: false
	PropagateSpawnedNodeChanges bool   // default: false (if true, changes need to be received through the corresponding channel)(automated by dashboard)
	InfoLoggerPath              string // *optional*
	InternalLoggerPath          string // *optional*
	WarningLoggerPath           string // *optional*
	InternalWarningLoggerPath   string // *optional*
	ErrorLoggerPath             string // *optional*
	DebugLoggerPath             string // *optional*
	Mailer                      *Mailer

	ResolverEndpoint     *TcpEndpoint // *required*
	BrokerConfigEndpoint *TcpEndpoint // *required*
}

func UnmarshalSpawner(data string) *Spawner {
	var spawner Spawner
	json.Unmarshal([]byte(data), &spawner)
	return &spawner
}

// Server applies to both http and websocket besides the fact that websocket is hardcoded to port 18251
type Dashboard struct {
	Server                         *TcpServer // *required*
	AutoStart                      bool       // default: false
	AddDashboardToDashboard        bool       // default: false
	HeapUpdateIntervalMs           uint64     // default: 0 = disabled
	GoroutineUpdateIntervalMs      uint64     // default: 0 = disabled
	NodeStatusIntervalMs           uint64     // default: 0 = disabled
	NodeSystemgeCounterIntervalMs  uint64     // default: 0 = disabled
	NodeHTTPCounterIntervalMs      uint64     // default: 0 = disabled
	NodeWebsocketCounterIntervalMs uint64     // default: 0 = disabled
	NodeBrokerCounterIntervalMs    uint64     // default: 0 = disabled
	NodeResolverCounterIntervalMs  uint64     // default: 0 = disabled
	NodeSpawnerCounterIntervalMs   uint64     // default: 0 = disabled
}

func UnmarshalDashboard(data string) *Dashboard {
	var dashboard Dashboard
	json.Unmarshal([]byte(data), &dashboard)
	return &dashboard
}
