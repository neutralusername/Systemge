package Config

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Systemge struct {
	HandleMessagesSequentially bool `json:"handleMessagesSequentially"` // default: false

	BrokerSubscribeDelayMs    uint64 `json:"brokerSubscribeDelayMs"`    // default: 0 (delay after failed broker subscription attempt)
	TopicResolutionLifetimeMs int64  `json:"topicResolutionLifetimeMs"` // default: <0 = infinite; 0 = resolve every time; >0 = resolve every n ms
	SyncResponseTimeoutMs     uint64 `json:"syncResponseTimeoutMs"`     // default: 0 = sync messages are not supported
	TcpTimeoutMs              uint64 `json:"tcpTimeoutMs"`              // default: 0 = block forever
	MaxSubscribeAttempts      uint64 `json:"maxSubscribeAttempts"`      // default: 0 = infinite

	ResolverEndpoints []*TcpEndpoint `json:"resolverEndpoints"` // at least one *required* (node attempts to resolve topics with these resolvers in the provided order. stops on first successful resolution)
}

func UnmarshalSystemge(data string) *Systemge {
	var systemge Systemge
	json.Unmarshal([]byte(data), &systemge)
	return &systemge
}

type Websocket struct {
	Pattern string     `json:"pattern"` // *required*
	Server  *TcpServer `json:"server"`  // *required*

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
	Server *TcpServer `json:"server"` // *required*
}

func UnmarshalHTTP(data string) *HTTP {
	var http HTTP
	json.Unmarshal([]byte(data), &http)
	return &http
}

type Broker struct {
	Server       *TcpServer   `json:"server"`       // *required*
	Endpoint     *TcpEndpoint `json:"endpoint"`     // *required*
	ConfigServer *TcpServer   `json:"configServer"` // *required*

	SyncTopics  []string `json:"syncTopics"`  // *required*
	AsyncTopics []string `json:"asyncTopics"` // *required*

	ResolverConfigEndpoints []*TcpEndpoint `json:"resolverConfigEndpoints"` // *optional* (propagates topics/endpoint to these resolvers)

	SyncResponseTimeoutMs uint64 `json:"syncResponseTimeoutMs"` // default: 0
	TcpTimeoutMs          uint64 `json:"tcpTimeoutMs"`          // default: 0 = block forever

	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 = unlimited
	MaxOriginSize            int    `json:"maxOriginSize"`            // default: 0 = unlimited
	MaxPayloadSize           int    `json:"maxPayloadSize"`           // default: 0 = unlimited
	MaxTopicSize             int    `json:"maxTopicSize"`             // default: 0 = unlimited
	MaxSyncKeySize           int    `json:"maxSyncKeySize"`           // default: 0 = unlimited
}

func UnmarshalBroker(data string) *Broker {
	var broker Broker
	json.Unmarshal([]byte(data), &broker)
	return &broker
}

type Resolver struct {
	Server       *TcpServer `json:"server"`       // *required*
	ConfigServer *TcpServer `json:"configServer"` // *required*

	TcpTimeoutMs     uint64                  `json:"tcpTimeoutMs"`     // default: 0 = block forever
	TopicResolutions map[string]*TcpEndpoint `json:"topicResolutions"` // *optional* (resolves these topics with the provided endpoints) (can get updated/overwritten by brokers)

	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 = unlimited
	MaxPayloadSize           int    `json:"maxPayloadSize"`           // default: 0 = unlimited
	MaxOriginSize            int    `json:"maxOriginSize"`            // default: 0 = unlimited
	MaxTopicSize             int    `json:"maxTopicSize"`             // default: 0 = unlimited
}

func UnmarshalResolver(data string) *Resolver {
	var resolver Resolver
	json.Unmarshal([]byte(data), &resolver)
	return &resolver
}
