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

type NewNode struct {
	NodeConfig      *Node      `json:"node"`            // *required*
	SystemgeConfig  *Systemge  `json:"systegmeConfig"`  // *optional*
	HttpConfig      *HTTP      `json:"httpConfig"`      // *optional*
	WebsocketConfig *Websocket `json:"websocketConfig"` // *optional*
}

func UnmarshalNode(data string) *Node {
	var node Node
	json.Unmarshal([]byte(data), &node)
	return &node
}
