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
	var tcpServer TcpServer
	json.Unmarshal([]byte(data), &tcpServer)
	return &tcpServer
}

type TcpEndpoint struct {
	Address string `json:"address"` // *required*
	Domain  string `json:"domain"`  // *optional*
	TlsCert string `json:"tlsCert"` // *optional* cert, NOT path!
}

func UnmarshalTcpEndpoint(data string) *TcpEndpoint {
	var tcpEndpoint TcpEndpoint
	json.Unmarshal([]byte(data), &tcpEndpoint)
	return &tcpEndpoint
}
