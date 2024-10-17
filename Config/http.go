package Config

import "encoding/json"

type HTTPServer struct {
	TcpServerConfig *TcpServer `json:"tcpServerConfig"` // *required*

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
