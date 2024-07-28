package Config

import "encoding/json"

type Node struct {
	Name string // *required*

	Mailer                    *Mailer
	InfoLoggerPath            string // *optional*
	InternalInfoLoggerPath    string // *optional*
	WarningLoggerPath         string // *optional*
	InternalWarningLoggerPath string // *optional*
	ErrorLoggerPath           string // *optional*
	DebugLoggerPath           string // *optional*

	RandomizerSeed int64 // default: 0
}

func UnmarshalNode(data string) *Node {
	var node Node
	json.Unmarshal([]byte(data), &node)
	return &node
}
