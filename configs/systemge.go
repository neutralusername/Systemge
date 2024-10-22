package configs

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type WebsocketListener struct {
	TcpListenerConfig *TcpListener `json:"tcpServerConfig"` // *required*
	Pattern           string       `json:"pattern"`         // *required* (the pattern that the underlying http server will listen to) (e.g. "/ws")

	Upgrader *websocket.Upgrader `json:"upgrader"` // *required*

	UpgradeRequestTimeoutMs uint32 `json:"upgradeRequestTimeoutMs"` // default: 0 (no timeout)
}

func UnmarshalWebsocketListener(data string) *WebsocketListener {
	var ws WebsocketListener
	err := json.Unmarshal([]byte(data), &ws)
	if err != nil {
		return nil
	}
	return &ws
}
