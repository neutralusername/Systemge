package Config

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Systemge struct {
	HandleMessagesSequentially bool `json:"handleMessagesSequentially"` // default: false (if true, the server will handle all incoming messages sequentially, which means only one message will be processed at a time) (if false, the server will handle messages concurrently)

	SyncRequestTimeoutMs            uint64 `json:"syncRequestTimeout"`              // default: 0 = infinite, which means SyncRequestChannel's need to be closed manually by the application or else there will be a memory leak
	SyncResponseLimit               uint64 `json:"syncResponseLimit"`               // default: 0 == sync responses are disabled
	TcpTimeoutMs                    uint64 `json:"tcpTimeoutMs"`                    // default: 0 = block forever
	MaxConnectionAttempts           uint64 `json:"maxConnectionAttempts"`           // default: 0 = infinite
	ConnectionAttemptDelayMs        uint64 `json:"connectionAttemptDelay"`          // default: 0 (delay after failed connection attempt)
	StopAfterOutgoingConnectionLoss bool   `json:"stopAfterOutgoingConnectionLoss"` // default: false (relevant if maxConnectionAttempts is set)

	ServerConfig    *TcpServer     `json:"serverConfig"`    // *required* (the configuration of this node's server)
	Endpoint        *TcpEndpoint   `json:"endpoint"`        // *optional* (the configuration of this node's endpoint) (can be shared with other nodes to let them connect during runtime)
	EndpointConfigs []*TcpEndpoint `json:"endpointConfigs"` // *required* (endpoint to other node's servers) (on startup, this node will attempt to establish connection to these endpoints)

	OutgoingConnectionRateLimiterBytes *RateLimiter `json:"outgoingConnectionRateLimiterBytes"` // *optional* (rate limiter for outgoing connections)
	OutgoingConnectionRateLimiterMsgs  *RateLimiter `json:"outgoingConnectionRateLimiterMsgs"`  // *optional* (rate limiter for outgoing connections)

	IncomingConnectionRateLimiterBytes *RateLimiter `json:"incomingConnectionRateLimiterBytes"` // *optional* (rate limiter for incoming connections)
	IncomingConnectionRateLimiterMsgs  *RateLimiter `json:"incomingConnectionRateLimiterMsgs"`  // *optional* (rate limiter for incoming connections)

	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 = unlimited (connections that attempt to send messages larger than this will be disconnected)
	MaxPayloadSize           int    `json:"maxPayloadSize"`           // default: <=0 = unlimited (messages that exceed this limit will be skipped)
	MaxTopicSize             int    `json:"maxTopicSize"`             // default: <=0 = unlimited (messages that exceed this limit will be skipped)
	MaxSyncTokenSize         int    `json:"maxSyncTokenSize"`         // default: <=0 = unlimited (messages that exceed this limit will be skipped)
	MaxNodeNameSize          uint64 `json:"maxNodeNameSize"`          // default: <=0 = unlimited (connections that attempt to send a node name larger than this will be rejected)
}

func UnmarshalSystemge(data string) *Systemge {
	var systemge Systemge
	json.Unmarshal([]byte(data), &systemge)
	return &systemge
}

type Websocket struct {
	Pattern      string     `json:"pattern"`      // *required* (the pattern that the underlying http server will listen to)
	ServerConfig *TcpServer `json:"serverConfig"` // *required* (the configuration of the underlying http server)

	ClientRateLimiterBytes *RateLimiter `json:"connectionRateLimiterBytes"` // *optional* (rate limiter for websocket clients)
	ClientRateLimiterMsgs  *RateLimiter `json:"connectionRateLimiterMsgs"`  // *optional* (rate limiter for websocket clients)

	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 = unlimited (connections that attempt to send messages larger than this will be disconnected)

	HandleClientMessagesSequentially bool   `json:"handleClientMessagesSequentially"` // default: false (if true, the server will handle messages from the same client sequentially)
	ClientMessageCooldownMs          uint64 `json:"clientMessageCooldownMs"`          // default: 0 (if a client sends messages faster than this, the server will ignore)
	ClientWatchdogTimeoutMs          uint64 `json:"clientWatchdogTimeoutMs"`          // default: 0 (if a client does not send a heartbeat message within this time, the server will disconnect the client)

	Upgrader *websocket.Upgrader `json:"upgrader"` // *required*

}

func UnmarshalWebsocket(data string) *Websocket {
	var websocket Websocket
	json.Unmarshal([]byte(data), &websocket)
	return &websocket
}

type HTTP struct {
	ServerConfig        *TcpServer `json:"serverConfig"`        // *required*
	MaxHeaderBytes      int        `json:"maxHeaderBytes"`      // default: 0 == 1 MB
	ReadHeaderTimeoutMs int        `json:"readHeaderTimeoutMs"` // default: 0 (no timeout)
	WriteTimeoutMs      int        `json:"writeTimeoutMs"`      // default: 0 (no timeout)
	MaxBodyBytes        int64      `json:"maxBodyBytes"`        // default: 0 (no limit)
}

func UnmarshalHTTP(data string) *HTTP {
	var http HTTP
	json.Unmarshal([]byte(data), &http)
	return &http
}
