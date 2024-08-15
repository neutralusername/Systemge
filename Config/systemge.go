package Config

type SystemgeReceiver struct {
	Name string `json:"name"` // *required*

	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*

	ProcessSequentially bool `json:"processSequentially"` // default: false (if true, the receiver will handle messages sequentially)
	ProcessChannelSize  int  `json:"processChannelSize"`  // default: 0 == no guarantee on order of arrival (if >0, the order of arrival is guaranteed as long as the channel is never full)

	RateLimiterBytes    *TokenBucketRateLimiter `json:"outgoingConnectionRateLimiterBytes"` // *optional* (rate limiter for outgoing connections)
	RateLimiterMessages *TokenBucketRateLimiter `json:"outgoingConnectionRateLimiterMsgs"`  // *optional* (rate limiter for outgoing connections)

	MaxPayloadSize   int `json:"maxPayloadSize"`   // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxTopicSize     int `json:"maxTopicSize"`     // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxSyncTokenSize int `json:"maxSyncTokenSize"` // default: <=0 == unlimited (messages that exceed this limit will be skipped)
}

type SystemgeServer struct {
	Name string `json:"name"` // *required*

	ListenerConfig *TcpListener `json:"serverConfig"` // *required* (the configuration of this node's server)
	Endpoint       *TcpEndpoint `json:"endpoint"`     // *optional* (the configuration of this node's endpoint) (can be shared with other nodes to let them connect during runtime)

	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*

	IpRateLimiter *IpRateLimiter `json:"ipRateLimiter"` // *optional* (rate limiter for incoming connections) (allows to limit the number of incoming connection attempts from the same IP) (it is more efficient to use a firewall for this purpose)

	MaxClientNameLength uint64 `json:"maxClientNameLength"` // default: 0 == unlimited (clients that attempt to send a name larger than this will be rejected)
}

type SystemgeConnection struct {
	Name string `json:"name"` // *required*

	RandomizerSeed int64 `json:"randomizerSeed"` // *optional*

	SyncRequestTimeoutMs uint64 `json:"syncRequestTimeout"` // default: 0 == infinite, which means SyncRequestChannel's need to be closed manually by the application or else there will be a memory leak
	TcpReceiveTimeoutMs  uint64 `json:"tcpReceiveTimeout"`  // default: 0 == block forever
	TcpSendTimeoutMs     uint64 `json:"tcpSendTimeout"`     // default: 0 == block forever

	TcpBufferBytes           uint32 `json:"tcpBufferBytes"`           // default: 0 == default (4KB)
	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 == unlimited (connections that attempt to send messages larger than this will be disconnected)
}

/* type SystemgeServer struct {
	Name string `json:"name"` // *required*

	ServerConfig *TcpListener `json:"serverConfig"` // *required* (the configuration of this node's server)
	Endpoint     *TcpEndpoint `json:"endpoint"`     // *optional* (the configuration of this node's endpoint) (can be shared with other nodes to let them connect during runtime)

	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*

	TcpTimeoutMs                                uint64 `json:"tcpTimeoutMs"`                                // default: 0 == block forever
	ProcessMessagesOfEachConnectionSequentially bool   `json:"processMessagesOfEachConnectionSequentially"` // default: false (if true, the server will handle messages from the same incoming connection sequentially) (if >1 incomming connection, multiple message handlers may run concurrently)
	ProcessAllMessagesSequentially              bool   `json:"processAllMessagesSequentially"`              // default: false (overrides ProcessMessagesOfEachConnectionSequentially) (guarantees, that only one message handler runs at a time)
	ProcessAllMessagesSequentiallyChannelSize   int    `json:"processAllMessagesSequentiallyChannelSize"`   // default: 0 == no guarantee on order of arrival (if >0, the order of arrival is guaranteed as long as the channel is never full (does NOT guarantee that messages arrive in the same chronological order as they were sent (technically not possible with multiple incoming connections and without a global clock)))

	RateLimterBytes     *TokenBucketRateLimiter `json:"incomingConnectionRateLimiterBytes"` // *optional* (rate limiter for incoming connections)
	RateLimiterMessages *TokenBucketRateLimiter `json:"incomingConnectionRateLimiterMsgs"`  // *optional* (rate limiter for incoming connections)

	TcpBufferBytes         uint32 `json:"tcpBufferBytes"`           // default: 0 == default (4KB)
	ClientMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 == unlimited (connections that attempt to send messages larger than this will be disconnected)
	MaxPayloadSize         int    `json:"maxPayloadSize"`           // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxTopicSize           int    `json:"maxTopicSize"`             // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxSyncTokenSize       int    `json:"maxSyncTokenSize"`         // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxClientNameLength    uint64 `json:"maxClientNameLength"`      // default: 0 == unlimited (clients that attempt to send a name larger than this will be rejected)
}

func UnmarshalSystemgeServer(data string) *SystemgeServer {
	var systemge SystemgeServer
	json.Unmarshal([]byte(data), &systemge)
	return &systemge
}

type SystemgeClient struct {
	Name string `json:"name"` // *required*

	EndpointConfig    *TcpEndpoint `json:"endpointConfig"`    // *required*
	MailerConfig      *Mailer      `json:"mailerConfig"`      // *optional*
	InfoLoggerPath    string       `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string       `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string       `json:"errorLoggerPath"`   // *optional*

	MaxConnectionAttempts      uint64 `json:"maxConnectionAttempts"`      // default: 0 == infinite
	ConnectionAttemptDelayMs   uint64 `json:"connectionAttemptDelay"`     // default: 0 (delay after failed connection attempt)
	RestartAfterConnectionLoss bool   `json:"restartAfterConnectionLoss"` // default: false (relevant if maxConnectionAttempts is set)
}

func UnmarshalSystemgeClient(data string) *SystemgeConnection {
	var systemge SystemgeConnection
	json.Unmarshal([]byte(data), &systemge)
	return &systemge
}
*/
