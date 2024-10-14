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

	Upgrader *websocket.Upgrader `json:"upgrader"` // *required*

	MaxSimultaneousAccepts  uint32 `json:"maxSimultaneousAccepts"`  // default: 1 (>0)
	UpgradeRequestTimeoutMs uint32 `json:"upgradeRequestTimeoutMs"` // default: 0 (no timeout)
}

func UnmarshalWebsocketListener(data string) *WebsocketListener {
	var ws WebsocketListener
	err := json.Unmarshal([]byte(data), &ws)
	if err != nil {
		return nil
	}
	return &ws
}

type WebsocketServer struct {
	WebsocketListenerConfig *WebsocketListener `json:"websocketListenerConfig"`   // *required*
	WebsocketClientConfig   *WebsocketClient   `json:"websocketConnectionConfig"` // *required*

	SessionManagerConfig *SessionManager `json:"sessionManagerConfig"` // *required*

	HandleClientsSequentially  bool `json:"handleClientsSequentially"`  // default: false
	HandleMessagesSequentially bool `json:"handleMessagesSequentially"` // default: false

	AcceptTimeoutMs uint32 `json:"acceptTimeoutMs"` // default: 0 (no timeout)
	ReadTimeoutMs   uint32 `json:"readTimeoutMs"`   // default: 0 (no timeout)
	WriteTimeoutMs  uint64 `json:"writeTimeoutMs"`  // default: 0 (no timeout)
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
