package configs

import "encoding/json"

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
	Address string `json:"address"` // *required* (e.g. "127.0.0.1:60009")
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
