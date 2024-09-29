package Config

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"golang.org/x/oauth2"
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

type WebsocketServer struct {
	TcpServerConfig *TcpServer `json:"tcpServerConfig"` // *required*
	Pattern         string     `json:"pattern"`         // *required* (the pattern that the underlying http server will listen to) (e.g. "/ws")

	RandomizerSeed int64 `json:"randomizerSeed"` // *optional*

	IpRateLimiter *IpRateLimiter `json:"ipRateLimiter"` // *optional* (rate limiter for incoming connections) (allows to limit the number of incoming connection attempts from the same IP) (it is more efficient to use a firewall for this purpose)

	WebsocketConnectionRateLimiterBytes    *TokenBucketRateLimiter `json:"websocketConnectionRateLimiterBytes"`    // *optional* (rate limiter for websocketConnections)
	WebsocketConnectionRateLimiterMessages *TokenBucketRateLimiter `json:"websocketConnectionRateLimiterMessages"` // *optional* (rate limiter for websocketConnections)

	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 = unlimited (connections that attempt to send messages larger than this will be disconnected)
	MaxPayloadSize           int    `json:"maxPayloadSize"`           // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxTopicSize             int    `json:"maxTopicSize"`             // default: <=0 == unlimited (messages that exceed this limit will be skipped)

	HandleMessageReceptionSequentially bool `json:"handleMessageReceptionSequentially"` // default: false (if true, the server will handle messages from the same websocketConnection sequentially)
	PropagateMessageHandlerWarnings    bool `json:"propagateMessageHandlerWarnings"`    // default: false (if true, the server will propagate warnings from message handlers to the websocketConnection)
	PropagateMessageHandlerErrors      bool `json:"propagateMessageHandlerErrors"`      // default: false (if true, the server will propagate errors from message handlers to the websocketConnection)

	ServerReadDeadlineMs int `json:"serverReadDeadlineMs"` // default: 60000 (1 minute, the server will disconnect websocketConnections that do not send messages within this time)

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

type Oauth2 struct {
	TcpServerConfig            *TcpServer     `json:"tcpServerConfig"`            // *required*
	RandomizerSeed             int64          `json:"randomizerSeed"`             // default: 0
	AuthPath                   string         `json:"authPath"`                   // *required*
	AuthCallbackPath           string         `json:"authCallbackPath"`           // *required*
	OAuth2Config               *oauth2.Config `json:"oAuth2Config"`               // *required*
	AuthRedirectUrl            string         `json:"authRedirectUrl"`            // *optional*
	CallbackFailureRedirectUrl string         `json:"callbackFailureRedirectUrl"` // *required*
	Oauth2State                string         `json:"oauth2State"`                // *required*
}

func UnmarshalOauth2(data string) *Oauth2 {
	var oauth2 Oauth2
	err := json.Unmarshal([]byte(data), &oauth2)
	if err != nil {
		return nil
	}
	return &oauth2
}
