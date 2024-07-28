package Config

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Systemge struct {
	HandleMessagesSequentially bool // default: false

	BrokerSubscribeDelayMs    uint64 // default: 0 (delay after failed broker subscription attempt)
	TopicResolutionLifetimeMs int64  // default: <0 = infinite; 0 = resolve every time; >0 = resolve every n ms
	SyncResponseTimeoutMs     uint64 // default: 0 = sync messages are not supported
	TcpTimeoutMs              uint64 // default: 0 = block forever
	MaxSubscribeAttempts      uint64 // default: 0 = infinite

	ResolverEndpoint *TcpEndpoint // *required*
}

func UnmarshalSystemge(data string) *Systemge {
	var systemge Systemge
	json.Unmarshal([]byte(data), &systemge)
	return &systemge
}

type Websocket struct {
	Pattern string     // *required*
	Server  *TcpServer // *required*

	HandleClientMessagesSequentially bool   // default: false
	ClientMessageCooldownMs          uint64 // default: 0
	ClientWatchdogTimeoutMs          uint64 // default: 0

	Upgrader *websocket.Upgrader // *required*

}

func UnmarshalWebsocket(data string) *Websocket {
	var websocket Websocket
	json.Unmarshal([]byte(data), &websocket)
	return &websocket
}

type HTTP struct {
	Server *TcpServer // *required*
}

func UnmarshalHTTP(data string) *HTTP {
	var http HTTP
	json.Unmarshal([]byte(data), &http)
	return &http
}

type Broker struct {
	Server       *TcpServer   // *required*
	Endpoint     *TcpEndpoint // *required*
	ConfigServer *TcpServer   // *required*

	SyncTopics  []string
	AsyncTopics []string

	ResolverConfigEndpoint *TcpEndpoint // *required*

	SyncResponseTimeoutMs uint64 // default: 0
	TcpTimeoutMs          uint64 // default: 0 = block forever

	IncomingMessageByteLimit uint64 // default: 0 = unlimited
	MaxOriginSize            int    // default: 0 = unlimited
	MaxPayloadSize           int    // default: 0 = unlimited
	MaxTopicSize             int    // default: 0 = unlimited
	MaxSyncKeySize           int    // default: 0 = unlimited
}

func UnmarshalBroker(data string) *Broker {
	var broker Broker
	json.Unmarshal([]byte(data), &broker)
	return &broker
}

type Resolver struct {
	Server       *TcpServer // *required*
	ConfigServer *TcpServer // *required*

	TcpTimeoutMs uint64 // default: 0 = block forever

	IncomingMessageByteLimit uint64 // default: 0 = unlimited
	MaxPayloadSize           int    // default: 0 = unlimited
	MaxOriginSize            int    // default: 0 = unlimited
	MaxTopicSize             int    // default: 0 = unlimited
}

func UnmarshalResolver(data string) *Resolver {
	var resolver Resolver
	json.Unmarshal([]byte(data), &resolver)
	return &resolver
}
