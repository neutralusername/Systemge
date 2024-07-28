package Config

import "encoding/json"

type TcpServer struct {
	Port        uint16 // *required*
	TlsCertPath string // *optional*
	TlsKeyPath  string // *optional*

	Blacklist []string // *optional*
	Whitelist []string // *optional* (if empty, all IPs are allowed)
}

func UnmarshalTcpServer(data string) *TcpServer {
	var tcpServer TcpServer
	json.Unmarshal([]byte(data), &tcpServer)
	return &tcpServer
}

type TcpEndpoint struct {
	Address string // *required*
	Domain  string // *optional*
	TlsCert string // *optional*
}

func UnmarshalTcpEndpoint(data string) *TcpEndpoint {
	var tcpEndpoint TcpEndpoint
	json.Unmarshal([]byte(data), &tcpEndpoint)
	return &tcpEndpoint
}
