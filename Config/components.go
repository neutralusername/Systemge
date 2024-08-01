package Config

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Systemge struct {
	HandleMessagesSequentially bool `json:"handleMessagesSequentially"` // default: false

	SyncRequestTimeoutMs            uint64 `json:"syncRequestTimeout"`              // default: 0 = infinite, which means SyncRequestChannel's need to be closed manually by the application or else there will be a memory leak
	SyncResponseLimit               uint64 `json:"syncResponseLimit"`               // default: 0 == sync responses are disabled
	TcpTimeoutMs                    uint64 `json:"tcpTimeoutMs"`                    // default: 0 = block forever
	MaxConnectionAttempts           uint64 `json:"maxConnectionAttempts"`           // default: 0 = infinite
	ConnectionAttemptDelayMs        uint64 `json:"connectionAttemptDelay"`          // default: 0 (delay after failed connection attempt)
	StopAfterOutgoingConnectionLoss bool   `json:"stopAfterOutgoingConnectionLoss"` // default: false (relevant if maxConnectionAttempts is set)

	ServerConfig    *TcpServer     `json:"serverConfig"`    // *required*
	Endpoint        *TcpEndpoint   `json:"endpoint"`        // *optional* (endpoint of this node)
	EndpointConfigs []*TcpEndpoint `json:"endpointConfigs"` // *required* (nodes which shall receive systemge messages by this node) (on connection, they share which message topics they are interested in and only those are sent)

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
	Pattern      string     `json:"pattern"`      // *required*
	ServerConfig *TcpServer `json:"serverConfig"` // *required*

	HandleClientMessagesSequentially bool   `json:"handleClientMessagesSequentially"` // default: false
	ClientMessageCooldownMs          uint64 `json:"clientMessageCooldownMs"`          // default: 0
	ClientWatchdogTimeoutMs          uint64 `json:"clientWatchdogTimeoutMs"`          // default: 0

	Upgrader *websocket.Upgrader `json:"upgrader"` // *required*

}

func UnmarshalWebsocket(data string) *Websocket {
	var websocket Websocket
	json.Unmarshal([]byte(data), &websocket)
	return &websocket
}

type HTTP struct {
	ServerConfig *TcpServer `json:"serverConfig"` // *required*
}

func UnmarshalHTTP(data string) *HTTP {
	var http HTTP
	json.Unmarshal([]byte(data), &http)
	return &http
}
