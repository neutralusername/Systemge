package Config

import "encoding/json"

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
	NodeConfig      *Node      `json:"node"`            // *required*
	SystemgeConfig  *Systemge  `json:"systegmeConfig"`  // *optional*
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
