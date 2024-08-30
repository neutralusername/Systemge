package Config

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type HTTPServer struct {
	Name string `json:"name"` // *required*

	TcpListenerConfig *TcpListener `json:"tcpListenerConfig"` // *required*

	InfoLoggerPath    string `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string `json:"errorLoggerPath"`   // *optional*

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

type WebsocketServer struct {
	Name string `json:"name"` // *required*

	TcpListenerConfig *TcpListener `json:"tcpListenerConfig"` // *required*
	Pattern           string       `json:"pattern"`           // *required* (the pattern that the underlying http server will listen to) (e.g. "/ws")

	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*
	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
	RandomizerSeed    int64   `json:"randomizerSeed"`    // *optional*

	IpRateLimiter *IpRateLimiter `json:"ipRateLimiter"` // *optional* (rate limiter for incoming connections) (allows to limit the number of incoming connection attempts from the same IP) (it is more efficient to use a firewall for this purpose)

	ClientRateLimiterBytes    *TokenBucketRateLimiter `json:"clientRateLimiterBytes"`    // *optional* (rate limiter for clients)
	ClientRateLimiterMessages *TokenBucketRateLimiter `json:"clientRateLimiterMessages"` // *optional* (rate limiter for clients)

	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 = unlimited (connections that attempt to send messages larger than this will be disconnected)

	HandleClientMessagesSequentially bool   `json:"handleClientMessagesSequentially"` // default: false (if true, the server will handle messages from the same client sequentially)
	ClientWatchdogTimeoutMs          uint64 `json:"clientWatchdogTimeoutMs"`          // default: 0 (if a client does not send a heartbeat message within this time, the server will disconnect the client)

	Upgrader *websocket.Upgrader `json:"upgrader"` // *required*
}

func UnmarshalWebsocketServer(data string) *WebsocketServer {
	var ws WebsocketServer
	err := json.Unmarshal([]byte(data), &ws)
	if err != nil {
		return nil
	}
	return &ws
}
