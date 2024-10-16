package Config

import "encoding/json"

type SystemgeConnectionAttempt struct {
	MaxServerNameLength   int    `json:"maxServerNameLength"`      // default: 0 == unlimited (servers that attempt to send a name larger than this will be rejected)
	MaxConnectionAttempts uint32 `json:"maxConnectionAttempts"`    // default: 0 == unlimited (the maximum number of reconnection attempts, after which the client will stop trying to reconnect)
	RetryIntervalMs       uint32 `json:"connectionAttemptDelayMs"` // default: 0 == no delay (the delay between reconnection attempts)

	TcpClientConfig             *TcpClient             `json:"tcpClientConfig"`             // *required*
	TcpSystemgeConnectionConfig *TcpSystemgeConnection `json:"tcpSystemgeConnectionConfig"` // *required*
}

func UnmarshalSystemgeConnectionAttempt(data string) *SystemgeConnectionAttempt {
	var systemgeClient SystemgeConnectionAttempt
	err := json.Unmarshal([]byte(data), &systemgeClient)
	if err != nil {
		return nil
	}
	return &systemgeClient
}

type SystemgeClient struct {
	TcpSystemgeConnectionConfig *TcpSystemgeConnection `json:"tcpSystemgeConnectionConfig"` // *required*
	TcpClientConfigs            []*TcpClient           `json:"tcpClientConfigs"`

	SessionManagerConfig *SessionManager `json:"sessionManagerConfig"` // *required*

	MaxServerNameLength int `json:"maxServerNameLength"` // default: 0 == unlimited (servers that attempt to send a name larger than this will be rejected)

	AutoReconnectAttempts    bool   `json:"autoReconnectAttempts"`    // default: false (if true, the client will attempt to reconnect if the connection is lost)
	ConnectionAttemptDelayMs uint32 `json:"connectionAttemptDelayMs"` // default: 1000 (the delay between reconnection attempts in milliseconds)
	MaxConnectionAttempts    uint32 `json:"maxConnectionAttempts"`    // default: 0 == unlimited (the maximum number of reconnection attempts, after which the client will stop trying to reconnect)
}

func UnmarshalSystemgeClient(data string) *SystemgeClient {
	var systemgeClient SystemgeClient
	err := json.Unmarshal([]byte(data), &systemgeClient)
	if err != nil {
		return nil
	}
	return &systemgeClient
}

type Server struct {
	TcpSystemgeListenerConfig   *TcpSystemgeListener   `json:"tcpSystemgeListenerConfig"`   // *required*
	TcpSystemgeConnectionConfig *TcpSystemgeConnection `json:"tcpSystemgeConnectionConfig"` // *required*
	SessionManagerConfig        *SessionManager        `json:"sessionManagerConfig"`        // *required*
}

func UnmarshalSystemgeServer(data string) *Server {
	var systemgeServer Server
	err := json.Unmarshal([]byte(data), &systemgeServer)
	if err != nil {
		return nil
	}
	return &systemgeServer
}

type TcpSystemgeListener struct {
	TcpServerConfig *TcpServer `json:"tcpServerConfig"` // *required*
}

func UnmarshalTcpSystemgeListener(data string) *TcpSystemgeListener {
	var tcpSystemgeListener TcpSystemgeListener
	err := json.Unmarshal([]byte(data), &tcpSystemgeListener)
	if err != nil {
		return nil
	}
	return &tcpSystemgeListener
}

type TcpSystemgeConnection struct {
	TcpReceiveTimeoutNs      int64  `json:"tcpReceiveTimeoutMs"`      // default: 0 == block forever
	TcpSendTimeoutMs         uint64 `json:"tcpSendTimeoutMs"`         // default: 0 == block forever
	TcpBufferBytes           uint32 `json:"tcpBufferBytes"`           // default: 0 == default (4KB)
	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 == unlimited (connections that attempt to send messages larger than this will be disconnected)
}

func UnmarshalTcpSystemgeConnection(data string) *TcpSystemgeConnection {
	var tcpSystemgeConnection TcpSystemgeConnection
	err := json.Unmarshal([]byte(data), &tcpSystemgeConnection)
	if err != nil {
		return nil
	}
	return &tcpSystemgeConnection
}
