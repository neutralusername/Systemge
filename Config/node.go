package Config

import (
	"encoding/json"

	"golang.org/x/oauth2"
)

type Node struct {
	Name string // *required*

	MailerConfig              *Mailer `json:"mailerConfig"`              // *optional*
	InfoLoggerPath            string  `json:"infoLoggerPath"`            // *optional*
	InternalInfoLoggerPath    string  `json:"internalInfoLoggerPath"`    // *optional*
	WarningLoggerPath         string  `json:"warningLoggerPath"`         // *optional*
	InternalWarningLoggerPath string  `json:"internalWarningLoggerPath"` // *optional*
	ErrorLoggerPath           string  `json:"errorLoggerPath"`           // *optional*
	DebugLoggerPath           string  `json:"debugLoggerPath"`           // *optional

	RandomizerSeed int64 `json:"randomizerSeed"` // *optional*
}

func UnmarshalNode(data string) *Node {
	var node Node
	json.Unmarshal([]byte(data), &node)
	return &node
}

type NewNode struct {
	NodeConfig      *Node      `json:"nodeConfig"`      // *required*
	SystemgeConfig  *Systemge  `json:"systemgeConfig"`  // *optional*
	HttpConfig      *HTTP      `json:"httpConfig"`      // *optional*
	WebsocketConfig *Websocket `json:"websocketConfig"` // *optional*
}

func UnmarshalNewNode(data string) *NewNode {
	var newNode NewNode
	json.Unmarshal([]byte(data), &newNode)
	return &newNode
}

// Server applies to both http and websocket besides the fact that websocket is hardcoded to port 18251
type Dashboard struct {
	NodeConfig   *Node      `json:"nodeConfig"`   // *required*
	ServerConfig *TcpServer `json:"serverConfig"` // *required*

	AutoStart                      bool   `json:"autoStart"`                      // default: false
	AddDashboardToDashboard        bool   `json:"addDashboardToDashboard"`        // default: false
	HeapUpdateIntervalMs           uint64 `json:"heapUpdateIntervalMs"`           // default: 0 = disabled
	GoroutineUpdateIntervalMs      uint64 `json:"goroutineUpdateIntervalMs"`      // default: 0 = disabled
	NodeStatusIntervalMs           uint64 `json:"nodeStatusIntervalMs"`           // default: 0 = disabled
	NodeSystemgeCounterIntervalMs  uint64 `json:"nodeSystemgeCounterIntervalMs"`  // default: 0 = disabled
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
	ServerConfig               *TcpServer                                                                  `json:"serverConfig"`               // *required*
	NodeConfig                 *Node                                                                       `json:"nodeConfig"`                 // *required*
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
