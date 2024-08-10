package Config

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Systemge struct {
	ProcessMessagesOfEachConnectionSequentially bool `json:"processMessagesOfEachConnectionSequentially"` // default: false (if true, the server will handle messages from the same incoming connection sequentially) (if >1 incomming connection, multiple message handlers may run concurrently)
	ProcessAllMessagesSequentially              bool `json:"processAllMessagesSequentially"`              // default: false (overrides ProcessMessagesOfEachConnectionSequentially) (guarantees, that only one message handler runs at a time)
	ProcessAllMessagesSequentiallyChannelSize   int  `json:"processAllMessagesSequentiallyChannelSize"`   // default: 0 == no guarantee on order of arrival (if >0, the order of arrival is guaranteed as long as the channel is never full (does NOT guarantee that messages arrive in the same chronological order as they were sent (technically not possible with multiple incoming connections and without a global clock)))

	SyncRequestTimeoutMs            uint64 `json:"syncRequestTimeout"`              // default: 0 == infinite, which means SyncRequestChannel's need to be closed manually by the application or else there will be a memory leak
	TcpTimeoutMs                    uint64 `json:"tcpTimeoutMs"`                    // default: 0 == block forever
	MaxConnectionAttempts           uint64 `json:"maxConnectionAttempts"`           // default: 0 == infinite
	ConnectionAttemptDelayMs        uint64 `json:"connectionAttemptDelay"`          // default: 0 (delay after failed connection attempt)
	StopAfterOutgoingConnectionLoss bool   `json:"stopAfterOutgoingConnectionLoss"` // default: false (relevant if maxConnectionAttempts is set)

	ServerConfig    *TcpServer     `json:"serverConfig"`    // *required* (the configuration of this node's server)
	Endpoint        *TcpEndpoint   `json:"endpoint"`        // *optional* (the configuration of this node's endpoint) (can be shared with other nodes to let them connect during runtime)
	EndpointConfigs []*TcpEndpoint `json:"endpointConfigs"` // *required* (endpoint to other node's servers) (on startup, this node will attempt to establish connection to these endpoints)

	OutgoingConnectionRateLimiterBytes *RateLimiter `json:"outgoingConnectionRateLimiterBytes"` // *optional* (rate limiter for outgoing connections)
	OutgoingConnectionRateLimiterMsgs  *RateLimiter `json:"outgoingConnectionRateLimiterMsgs"`  // *optional* (rate limiter for outgoing connections)

	IncomingConnectionRateLimiterBytes *RateLimiter `json:"incomingConnectionRateLimiterBytes"` // *optional* (rate limiter for incoming connections)
	IncomingConnectionRateLimiterMsgs  *RateLimiter `json:"incomingConnectionRateLimiterMsgs"`  // *optional* (rate limiter for incoming connections)

	TcpBufferBytes           uint32 `json:"tcpBufferBytes"`           // default: 0 == default (4KB)
	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 == unlimited (connections that attempt to send messages larger than this will be disconnected)
	MaxPayloadSize           int    `json:"maxPayloadSize"`           // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxTopicSize             int    `json:"maxTopicSize"`             // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxSyncTokenSize         int    `json:"maxSyncTokenSize"`         // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxNodeNameSize          uint64 `json:"maxNodeNameSize"`          // default: 0 == unlimited (connections that attempt to send a node name larger than this will be rejected)
}

func UnmarshalSystemge(data string) *Systemge {
	var systemge Systemge
	json.Unmarshal([]byte(data), &systemge)
	return &systemge
}

type Websocket struct {
	Pattern      string     `json:"pattern"`      // *required* (the pattern that the underlying http server will listen to) (e.g. "/ws")
	ServerConfig *TcpServer `json:"serverConfig"` // *required* (the configuration of the underlying http server)

	ClientRateLimiterBytes *RateLimiter `json:"connectionRateLimiterBytes"` // *optional* (rate limiter for websocket clients)
	ClientRateLimiterMsgs  *RateLimiter `json:"connectionRateLimiterMsgs"`  // *optional* (rate limiter for websocket clients)

	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 = unlimited (connections that attempt to send messages larger than this will be disconnected)

	HandleClientMessagesSequentially bool   `json:"handleClientMessagesSequentially"` // default: false (if true, the server will handle messages from the same client sequentially)
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
