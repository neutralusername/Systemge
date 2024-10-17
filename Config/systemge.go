package Config

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type SystemgeConnectionAttempt struct {
	MaxServerNameLength   int    `json:"maxServerNameLength"`      // default: 0 == unlimited (servers that attempt to send a name larger than this will be rejected)
	MaxConnectionAttempts uint32 `json:"maxConnectionAttempts"`    // default: 0 == unlimited (the maximum number of reconnection attempts, after which the client will stop trying to reconnect)
	RetryIntervalMs       uint32 `json:"connectionAttemptDelayMs"` // default: 0 == no delay (the delay between reconnection attempts)

	TcpClientConfig             *TcpClient     `json:"tcpClientConfig"`             // *required*
	TcpSystemgeConnectionConfig *TcpConnection `json:"tcpSystemgeConnectionConfig"` // *required*
}

func UnmarshalSystemgeConnectionAttempt(data string) *SystemgeConnectionAttempt {
	var systemgeClient SystemgeConnectionAttempt
	err := json.Unmarshal([]byte(data), &systemgeClient)
	if err != nil {
		return nil
	}
	return &systemgeClient
}

type TcpListener struct {
	TcpServerConfig *TcpServer `json:"tcpServerConfig"` // *required*
}

func UnmarshalTcpSystemgeListener(data string) *TcpListener {
	var tcpSystemgeListener TcpListener
	err := json.Unmarshal([]byte(data), &tcpSystemgeListener)
	if err != nil {
		return nil
	}
	return &tcpSystemgeListener
}

type TcpConnection struct {
	TcpReceiveTimeoutNs      int64  `json:"tcpReceiveTimeoutMs"`      // default: 0 == block forever
	TcpBufferBytes           uint32 `json:"tcpBufferBytes"`           // default: 0 == default (4KB)
	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 == unlimited (connections that attempt to send messages larger than this will be disconnected)
}

func UnmarshalTcpSystemgeConnection(data string) *TcpConnection {
	var tcpSystemgeConnection TcpConnection
	err := json.Unmarshal([]byte(data), &tcpSystemgeConnection)
	if err != nil {
		return nil
	}
	return &tcpSystemgeConnection
}

type WebsocketListener struct {
	TcpServerConfig *TcpServer `json:"tcpServerConfig"` // *required*
	Pattern         string     `json:"pattern"`         // *required* (the pattern that the underlying http server will listen to) (e.g. "/ws")

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
