package Config

import "encoding/json"

type TcpServer struct {
	Port        uint16 // *required*
	TlsCertPath string // *optional*
	TlsKeyPath  string // *optional*
}

type TcpEndpoint struct {
	Address string // *required*
	Domain  string // *optional*
	TlsCert string // *optional*
}

func (tcpEndpoint TcpEndpoint) Marshal() string {
	json, _ := json.Marshal(TcpEndpoint{
		Address: tcpEndpoint.Address,
		Domain:  tcpEndpoint.Domain,
		TlsCert: tcpEndpoint.TlsCert,
	})
	return string(json)
}

func UnmarshalTcpEndpoint(data string) *TcpEndpoint {
	var resolutionData TcpEndpoint
	json.Unmarshal([]byte(data), &resolutionData)
	endpoint := TcpEndpoint{
		Address: resolutionData.Address,
		Domain:  resolutionData.Domain,
		TlsCert: resolutionData.TlsCert,
	}
	return &endpoint
}
