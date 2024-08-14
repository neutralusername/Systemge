package Config

import "encoding/json"

type HTTP struct {
	ServerConfig        *TcpServer `json:"serverConfig"`        // *required*
	MaxHeaderBytes      int        `json:"maxHeaderBytes"`      // default: <=0 == 1 MB (whichever value you choose, golangs http package will add 4096 bytes on top of it....)
	ReadHeaderTimeoutMs int        `json:"readHeaderTimeoutMs"` // default: 0 (no timeout)
	WriteTimeoutMs      int        `json:"writeTimeoutMs"`      // default: 0 (no timeout)
	MaxBodyBytes        int64      `json:"maxBodyBytes"`        // default: 0 (no limit)
}

func UnmarshalHTTP(data string) *HTTP {
	var http HTTP
	json.Unmarshal([]byte(data), &http)
	return &http
}
