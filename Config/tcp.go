package Config

import "encoding/json"

type TcpServer struct {
	Port        uint16 `json:"port"`        // *required*
	TlsCertPath string `json:"tlsCertPath"` // *optional* cert path!
	TlsKeyPath  string `json:"tlsKeyPath"`  // *optional*

	Blacklist []string `json:"blacklist"` // *optional* (if empty, all IPs are allowed)
	Whitelist []string `json:"whitelist"` // *optional* (if empty, all IPs are allowed)
}

func UnmarshalTcpServer(data string) *TcpServer {
	var tcpListener TcpServer
	err := json.Unmarshal([]byte(data), &tcpListener)
	if err != nil {
		return nil
	}
	return &tcpListener
}

type TcpClient struct {
	Address string `json:"address"` // *required* (e.g. "127.0.0.1:60009")
	Domain  string `json:"domain"`  // *optional* (e.g. "example.com")
	TlsCert string `json:"tlsCert"` // *optional* cert, NOT path!
}

func UnmarshalTcpClient(data string) *TcpClient {
	var tcpEndpoint TcpClient
	err := json.Unmarshal([]byte(data), &tcpEndpoint)
	if err != nil {
		return nil
	}
	return &tcpEndpoint
}
