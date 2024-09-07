package Config

import "encoding/json"

type SystemgeClient struct {
	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*

	ConnectionConfig *TcpSystemgeConnection `json:"connectionConfig"` // *required*
	ClientConfigs    []*TcpClient           `json:"clientConfigs"`

	MaxServerNameLength int `json:"maxServerNameLength"` // default: 0 == unlimited (servers that attempt to send a name larger than this will be rejected)

	Reconnect                bool   `json:"reconnect"`               // default: false (if true, the client will attempt to reconnect if the connection is lost)
	ConnectionAttemptDelayMs uint64 `json:"reconnectAttemptDelayMs"` // default: 1000 (the delay between reconnection attempts in milliseconds)
	MaxConnectionAttempts    uint32 `json:"maxReconnectAttempts"`    // default: 0 == unlimited (the maximum number of reconnection attempts, after which the client will stop trying to reconnect)
}

func UnmarshalSystemgeClient(data string) *SystemgeClient {
	var systemgeClient SystemgeClient
	err := json.Unmarshal([]byte(data), &systemgeClient)
	if err != nil {
		return nil
	}
	return &systemgeClient
}

type SystemgeServer struct {
	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*

	ListenerConfig   *TcpSystemgeListener   `json:"listenerConfig"`   // *required*
	ConnectionConfig *TcpSystemgeConnection `json:"connectionConfig"` // *required*
}

func UnmarshalSystemgeServer(data string) *SystemgeServer {
	var systemgeServer SystemgeServer
	err := json.Unmarshal([]byte(data), &systemgeServer)
	if err != nil {
		return nil
	}
	return &systemgeServer
}

type TcpSystemgeListener struct {
	TcpServerConfig *TcpServer `json:"tcpServerConfig"` // *required*

	IpRateLimiter       *IpRateLimiter `json:"ipRateLimiter"`       // *optional* (rate limiter for incoming connections) (allows to limit the number of incoming connection attempts from the same IP) (it is more efficient to use a firewall for this purpose)
	MaxClientNameLength uint64         `json:"maxClientNameLength"` // default: 0 == unlimited (clients that attempt to send a name larger than this will be rejected)
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
	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*

	RandomizerSeed int64 `json:"randomizerSeed"` // *optional*

	SyncRequestTimeoutMs uint64 `json:"syncRequestTimeoutMs"` // default: 0 == infinite, which means SyncRequestChannel's need to be closed manually by the application or else there will be a memory leak
	TcpReceiveTimeoutMs  uint64 `json:"tcpReceiveTimeoutMs"`  // default: 0 == block forever
	TcpSendTimeoutMs     uint64 `json:"tcpSendTimeoutMs"`     // default: 0 == block forever

	TcpBufferBytes           uint32 `json:"tcpBufferBytes"`           // default: 0 == default (4KB)
	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 == unlimited (connections that attempt to send messages larger than this will be disconnected)

	ProcessingChannelCapacity int `json:"processingChannelSize"` // default: 0 (how many messages can be received before being processed (n+1))

	RateLimiterBytes    *TokenBucketRateLimiter `json:"rateLimiterBytes"`    // *optional*
	RateLimiterMessages *TokenBucketRateLimiter `json:"rateLimiterMessages"` // *optional*

	MaxPayloadSize   int `json:"maxPayloadSize"`   // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxTopicSize     int `json:"maxTopicSize"`     // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxSyncTokenSize int `json:"maxSyncTokenSize"` // default: <=0 == unlimited (messages that exceed this limit will be skipped)
}

func UnmarshalTcpSystemgeConnection(data string) *TcpSystemgeConnection {
	var tcpSystemgeConnection TcpSystemgeConnection
	err := json.Unmarshal([]byte(data), &tcpSystemgeConnection)
	if err != nil {
		return nil
	}
	return &tcpSystemgeConnection
}
