package Config

import (
	"encoding/json"

	"golang.org/x/oauth2"
)

// ServerConfig applies to both http and websocket besides the fact that websocket is hardcoded to port 18251
type Dashboard struct {
	NodeConfig                *Node      `json:"nodeConfig"`                // *required*
	ServerConfig              *TcpServer `json:"serverConfig"`              // *required*
	AutoStart                 bool       `json:"autoStart"`                 // default: false
	AddDashboardToDashboard   bool       `json:"addDashboardToDashboard"`   // default: false
	HeapUpdateIntervalMs      uint64     `json:"heapUpdateIntervalMs"`      // default: 0 = disabled
	GoroutineUpdateIntervalMs uint64     `json:"goroutineUpdateIntervalMs"` // default: 0 = disabled
	NodeStatusIntervalMs      uint64     `json:"nodeStatusIntervalMs"`      // default: 0 = disabled

	NodeSystemgeClientCounterIntervalMs             uint64 `json:"nodeSystemgeClientCounterIntervalMs"`             // default: 0 = disabled
	NodeSystemgeClientRateLimitCounterIntervalMs    uint64 `json:"nodeSystemgeClientRateLimitCounterIntervalMs"`    // default: 0 = disabled
	NodeSystemgeClientConnectionCounterIntervalMs   uint64 `json:"nodeSystemgeClientConnectionCounterIntervalMs"`   // default: 0 = disabled
	NodeSystemgeClientSyncResponseCounterIntervalMs uint64 `json:"nodeSystemgeClientSyncResponseCounterIntervalMs"` // default: 0 = disabled
	NodeSystemgeClientAsyncMessageCounterIntervalMs uint64 `json:"nodeSystemgeClientAsyncMessageCounterIntervalMs"` // default: 0 = disabled
	NodeSystemgeClientSyncRequestCounterIntervalMs  uint64 `json:"nodeSystemgeClientSyncRequestCounterIntervalMs"`  // default: 0 = disabled
	NodeSystemgeClientTopicCounterIntervalMs        uint64 `json:"nodeSystemgeClientTopicCounterIntervalMs"`        // default: 0 = disabled

	NodeSystemgeServerCounterIntervalMs             uint64 `json:"nodeSystemgeServerCounterIntervalMs"`             // default: 0 = disabled
	NodeSystemgeServerRateLimitCounterIntervalMs    uint64 `json:"nodeSystemgeServerRateLimitCounterIntervalMs"`    // default: 0 = disabled
	NodeSystemgeServerConnectionCounterIntervalMs   uint64 `json:"nodeSystemgeServerConnectionCounterIntervalMs"`   // default: 0 = disabled
	NodeSystemgeServerSyncResponseCounterIntervalMs uint64 `json:"nodeSystemgeServerSyncResponseCounterIntervalMs"` // default: 0 = disabled
	NodeSystemgeServerAsyncMessageCounterIntervalMs uint64 `json:"nodeSystemgeServerAsyncMessageCounterIntervalMs"` // default: 0 = disabled
	NodeSystemgeServerSyncRequestCounterIntervalMs  uint64 `json:"nodeSystemgeServerSyncRequestCounterIntervalMs"`  // default: 0 = disabled
	NodeSystemgeServerTopicCounterIntervalMs        uint64 `json:"nodeSystemgeServerTopicCounterIntervalMs"`        // default: 0 = disabled

	NodeHTTPCounterIntervalMs      uint64 `json:"nodeHTTPCounterIntervalMs"`      // default: 0 = disabled
	NodeWebsocketCounterIntervalMs uint64 `json:"nodeWebsocketCounterIntervalMs"` // default: 0 = disabled
	NodeSpawnerCounterIntervalMs   uint64 `json:"nodeSpawnerCounterIntervalMs"`   // default: 0 = disabled
}

func UnmarshalDashboard(data string) *Dashboard {
	var dashboard Dashboard
	json.Unmarshal([]byte(data), &dashboard)
	return &dashboard
}

type Oauth2 struct {
	NodeConfig   *Node      `json:"nodeConfig"`   // *required*
	ServerConfig *TcpServer `json:"serverConfig"` // *required*

	AuthPath                   string                                                                      `json:"authPath"`                   // *required*
	AuthCallbackPath           string                                                                      `json:"authCallbackPath"`           // *required*
	OAuth2Config               *oauth2.Config                                                              `json:"oAuth2Config"`               // *required*
	AuthRedirectUrl            string                                                                      `json:"authRedirectUrl"`            // *optional*
	CallbackSuccessRedirectUrl string                                                                      `json:"callbackSuccessRedirectUrl"` // *required*
	CallbackFailureRedirectUrl string                                                                      `json:"callbackFailureRedirectUrl"` // *required*
	TokenHandler               func(*oauth2.Config, *oauth2.Token) (string, map[string]interface{}, error) `json:"-"`
	SessionLifetimeMs          uint64                                                                      `json:"sessionLifetimeMs"` // default: 0
	Oauth2State                string                                                                      `json:"oauth2State"`       // *required*
}

func UnmarshalOauth2(data string) *Oauth2 {
	var oauth2 Oauth2
	json.Unmarshal([]byte(data), &oauth2)
	return &oauth2
}

type Spawner struct {
	NodeConfig     *Node           `json:"nodeConfig"`     // *required*
	SystemgeConfig *SystemgeServer `json:"systemgeConfig"` // *required*

	PropagateSpawnedNodeChanges bool `json:"propagateSpawnedNodeChanges"` // default: false (if true, changes need to be received through the corresponding channel) (automated by dashboard)
}

func UnmarshalSpawner(data string) *Spawner {
	var spawner Spawner
	json.Unmarshal([]byte(data), &spawner)
	return &spawner
}
