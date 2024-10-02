package Config

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type HTTPServer struct {
	TcpServerConfig *TcpServer `json:"tcpServerConfig"` // *required*

	HttpErrorPath string `json:"httpErrorPath"` // *optional* (logged to standard output if empty)

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

type WebsocketListener struct {
	TcpServerConfig *TcpServer `json:"tcpServerConfig"` // *required*
	Pattern         string     `json:"pattern"`         // *required* (the pattern that the underlying http server will listen to) (e.g. "/ws")

	IpRateLimiter *IpRateLimiter `json:"ipRateLimiter"` // *optional* (rate limiter for incoming connections) (allows to limit the number of incoming connection attempts from the same IP) (it is more efficient to use a firewall for this purpose)

	Upgrader *websocket.Upgrader `json:"upgrader"` // *required*
}

type WebsocketConnection struct {
	TopicManagerConfig *TopicManager `json:"topicManagerConfig"` // *required*

	ReadDeadlineMs int `json:"serverReadDeadlineMs"` // default: 60000 (1 minute, the server will disconnect websocketConnections that do not send messages within this time)

	MessageChannelCapacity uint32 `json:"messageChannelCapacity"` // default: 0 (how many messages can be received before being processed (n+1))

	HandleMessageReceptionSequentially bool `json:"handleMessageReceptionSequentially"` // default: false (if true, the server will handle messages from the same websocketConnection sequentially)

	RateLimiterBytes    *TokenBucketRateLimiter `json:"rateLimiterBytes"`    // *optional* (rate limiter for incoming messages) (allows to limit the number of incoming messages per second)
	RateLimiterMessages *TokenBucketRateLimiter `json:"rateLimiterMessages"` // *optional* (rate limiter for incoming messages) (allows to limit the number of incoming messages per second)

	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 = unlimited (connections that attempt to send messages larger than this will be disconnected)
	MaxTopicSize             int    `json:"maxTopicSize"`             // default: 0 = unlimited (topics that attempt to store more messages than this will be truncated)
	MaxPayloadSize           int    `json:"maxPayloadSize"`           // default: 0 = unlimited (payloads that attempt to store more messages than this will be truncated)

}

type WebsocketServer struct {
	WebsocketListenerConfig   *WebsocketListener   `json:"websocketListenerConfig"`   // *required*
	WebsocketConnectionConfig *WebsocketConnection `json:"websocketConnectionConfig"` // *required*
	SessionManagerConfig      *SessionManager      `json:"sessionManagerConfig"`      // *required*
}

func UnmarshalWebsocketServer(data string) *WebsocketServer {
	var ws WebsocketServer
	err := json.Unmarshal([]byte(data), &ws)
	if err != nil {
		return nil
	}
	return &ws
}

type SingleRequestClient struct {
	TcpSystemgeConnectionConfig *TcpSystemgeConnection `json:"tcpSystemgeConnectionConfig"` // *required*
	TcpClientConfig             *TcpClient             `json:"tcpClientConfig"`             // *required*
	MaxServerNameLength         int                    `json:"maxServerNameLength"`         // default: <=0 == unlimited (clients that attempt to send a name larger than this will be rejected)
}

func UnmarshalCommandClient(data string) *SingleRequestClient {
	var commandClient SingleRequestClient
	err := json.Unmarshal([]byte(data), &commandClient)
	if err != nil {
		return nil
	}
	return &commandClient
}

type SingleRequestServer struct {
	SystemgeServerConfig *SystemgeServer `json:"systemgeServerConfig"` // *required*
}

func UnmarshalSingleRequestServer(data string) *SingleRequestServer {
	var singleRequestServer SingleRequestServer
	err := json.Unmarshal([]byte(data), &singleRequestServer)
	if err != nil {
		return nil
	}
	return &singleRequestServer
}
