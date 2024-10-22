package configs

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type HTTPServer struct {
	TcpListenerConfig *TcpListener `json:"tcpServerConfig"` // *required*

	HttpErrorLogPath string `json:"httpErrorPath"` // *optional* (logged to standard output if empty)

	DelayNs             int64 `json:"delayNs"`             // default: 0 (no delay)
	MaxHeaderBytes      int   `json:"maxHeaderBytes"`      // default: <=0 == 1 MB (whichever value you choose, golangs http package will add 4096 bytes on top of it....)
	ReadHeaderTimeoutMs int   `json:"readHeaderTimeoutMs"` // default: 0 (no timeout)
	WriteTimeoutMs      int   `json:"writeTimeoutMs"`      // default: 0 (no timeout)
	MaxBodyBytes        int64 `json:"maxBodyBytes"`        // default: 0 (no limit)
}

func UnmarshalHTTPServer(data string) *HTTPServer {
	var http HTTPServer
	err := json.Unmarshal([]byte(data), &http)
	if err != nil {
		return nil
	}
	return &http
}

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

type TcpListener struct {
	Port        uint16 `json:"port"`        // *required*
	Domain      string `json:"domain"`      // *optional* (e.g. "example.com")
	TlsCertPath string `json:"tlsCertPath"` // *optional* cert path!
	TlsKeyPath  string `json:"tlsKeyPath"`  // *optional*
}

func UnmarshalTcpListener(data string) *TcpListener {
	var tcpListener TcpListener
	err := json.Unmarshal([]byte(data), &tcpListener)
	if err != nil {
		return nil
	}
	return &tcpListener
}

type TcpClient struct {
	Port    uint16 `json:"port"`    // *required*
	Domain  string `json:"domain"`  // *optional* (e.g. "example.com")
	TlsCert string `json:"tlsCert"` // *optional* cert, NOT path!
}

func UnmarshalTcpClient(data string) *TcpClient {
	var tcpClient TcpClient
	err := json.Unmarshal([]byte(data), &tcpClient)
	if err != nil {
		return nil
	}
	return &tcpClient
}
