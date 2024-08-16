package Config

type SystemgeClient struct {
	Name string `json:"name"` // *required*

	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*

	ConnectionConfig *SystemgeConnection `json:"connectionConfig"` // *required*
	ReceiverConfig   *SystemgeReceiver   `json:"receiverConfig"`   // *required*

	Reconnect               bool   `json:"reconnect"`               // default: false (if true, the client will attempt to reconnect if the connection is lost)
	ReconnectAttemptDelayMs uint64 `json:"reconnectAttemptDelayMs"` // default: 1000 (the delay between reconnection attempts in milliseconds)
	MaxReconnectAttempts    uint64 `json:"maxReconnectAttempts"`    // default: 0 == unlimited (the maximum number of reconnection attempts, after which the client will stop trying to reconnect)
}

type SystemgeServer struct {
	Name string `json:"name"` // *required*

	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*

	ConnectionConfig *SystemgeConnection `json:"connectionConfig"` // *required*
	ReceiverConfig   *SystemgeReceiver   `json:"receiverConfig"`   // *required*
}

type SystemgeReceiver struct {
	Name string `json:"name"` // *required*

	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*

	ProcessSequentially   bool `json:"processSequentially"`   // default: false (if true, the receiver will handle messages sequentially)
	ProcessingChannelSize int  `json:"processingChannelSize"` // default: 0 == no guarantee on order of arrival (irrelevant if ProcessSequentially is false) (if >0, the order of arrival is guaranteed as long as the channel is never full)

	RateLimiterBytes    *TokenBucketRateLimiter `json:"rateLimiterBytes"`    // *optional*
	RateLimiterMessages *TokenBucketRateLimiter `json:"rateLimiterMessages"` // *optional*

	MaxPayloadSize   int `json:"maxPayloadSize"`   // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxTopicSize     int `json:"maxTopicSize"`     // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxSyncTokenSize int `json:"maxSyncTokenSize"` // default: <=0 == unlimited (messages that exceed this limit will be skipped)
}

type SystemgeListener struct {
	ListenerConfig *TcpListener `json:"listenerConfig"` // *required*
	EndpointConfig *TcpEndpoint `json:"endpointConfig"` // *optional*

	IpRateLimiter       *IpRateLimiter `json:"ipRateLimiter"`       // *optional* (rate limiter for incoming connections) (allows to limit the number of incoming connection attempts from the same IP) (it is more efficient to use a firewall for this purpose)
	MaxClientNameLength uint64         `json:"maxClientNameLength"` // default: 0 == unlimited (clients that attempt to send a name larger than this will be rejected)
}

type SystemgeConnection struct {
	RandomizerSeed int64 `json:"randomizerSeed"` // *optional*

	SyncRequestTimeoutMs uint64 `json:"syncRequestTimeoutMs"` // default: 0 == infinite, which means SyncRequestChannel's need to be closed manually by the application or else there will be a memory leak
	TcpReceiveTimeoutMs  uint64 `json:"tcpReceiveTimeoutMs"`  // default: 0 == block forever
	TcpSendTimeoutMs     uint64 `json:"tcpSendTimeoutMs"`     // default: 0 == block forever

	TcpBufferBytes           uint32 `json:"tcpBufferBytes"`           // default: 0 == default (4KB)
	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 == unlimited (connections that attempt to send messages larger than this will be disconnected)
}
