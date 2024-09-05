package Config

import "encoding/json"

type TcpServer struct {
	Port        uint16 `json:"port"`        // *required*
	TlsCertPath string `json:"tlsCertPath"` // *optional* cert path!
	TlsKeyPath  string `json:"tlsKeyPath"`  // *optional*

	Blacklist []string `json:"blacklist"` // *optional* (if empty, all IPs are allowed)
	Whitelist []string `json:"whitelist"` // *optional* (if empty, all IPs are allowed)
}

func UnmarshalTcpServer(data string) *TcpServer {
	var tcpListener TcpServer
	err := json.Unmarshal([]byte(data), &tcpListener)
	if err != nil {
		return nil
	}
	return &tcpListener
}

type TcpClient struct {
	Address string `json:"address"` // *required* (e.g. "127.0.0.1:60009")
	Domain  string `json:"domain"`  // *optional* (e.g. "example.com")
	TlsCert string `json:"tlsCert"` // *optional* cert, NOT path!
}

func UnmarshalTcpClient(data string) *TcpClient {
	var tcpClient TcpClient
	err := json.Unmarshal([]byte(data), &tcpClient)
	if err != nil {
		return nil
	}
	return &tcpClient
}

type TcpListener struct {
	TcpServerConfig *TcpServer `json:"tcpServerConfig"` // *required*

	IpRateLimiter       *IpRateLimiter `json:"ipRateLimiter"`       // *optional* (rate limiter for incoming connections) (allows to limit the number of incoming connection attempts from the same IP) (it is more efficient to use a firewall for this purpose)
	MaxClientNameLength uint64         `json:"maxClientNameLength"` // default: 0 == unlimited (clients that attempt to send a name larger than this will be rejected)
}

type TcpConnection struct {
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
