package Config

/* type SingleRequestClient struct {
	TcpSystemgeConnectionConfig *TcpConnection `json:"tcpSystemgeConnectionConfig"` // *required*
	TcpClientConfig             *TcpClient     `json:"tcpClientConfig"`             // *required*
	MaxServerNameLength         int            `json:"maxServerNameLength"`         // default: <=0 == unlimited (clients that attempt to send a name larger than this will be rejected)
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
	SystemgeServerConfig *Server `json:"systemgeServerConfig"` // *required*
}

func UnmarshalSingleRequestServer(data string) *SingleRequestServer {
	var singleRequestServer SingleRequestServer
	err := json.Unmarshal([]byte(data), &singleRequestServer)
	if err != nil {
		return nil
	}
	return &singleRequestServer
}
*/

/*
type SystemgeClient struct {
	TcpSystemgeConnectionConfig *TcpConnection `json:"tcpSystemgeConnectionConfig"` // *required*
	TcpClientConfigs            []*TcpClient   `json:"tcpClientConfigs"`

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
*/
