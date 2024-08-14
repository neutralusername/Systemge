package Config

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type WebsocketServer struct {
	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*
	Mailer            *Mailer `json:"mailer"`            // *optional*
	RandomizerSeed    int64   `json:"randomizerSeed"`    // *optional*

	Pattern      string     `json:"pattern"`      // *required* (the pattern that the underlying http server will listen to) (e.g. "/ws")
	ServerConfig *TcpServer `json:"serverConfig"` // *required* (the configuration of the underlying http server)

	ClientRateLimiterBytes    *TokenBucketRateLimiter `json:"connectionRateLimiterBytes"` // *optional* (rate limiter for websocket clients)
	ClientRateLimiterMessages *TokenBucketRateLimiter `json:"connectionRateLimiterMsgs"`  // *optional* (rate limiter for websocket clients)

	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 = unlimited (connections that attempt to send messages larger than this will be disconnected)

	HandleClientMessagesSequentially bool   `json:"handleClientMessagesSequentially"` // default: false (if true, the server will handle messages from the same client sequentially)
	ClientWatchdogTimeoutMs          uint64 `json:"clientWatchdogTimeoutMs"`          // default: 0 (if a client does not send a heartbeat message within this time, the server will disconnect the client)

	Upgrader *websocket.Upgrader `json:"upgrader"` // *required*

}

func UnmarshalWebsocketServer(data string) *WebsocketServer {
	var websocket WebsocketServer
	json.Unmarshal([]byte(data), &websocket)
	return &websocket
}
