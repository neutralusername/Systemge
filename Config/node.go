package Config

import "encoding/json"

type SystemgeClient struct {
	Name string `json:"name"` // *required*

	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*

	SyncRequestTimeoutMs            uint64 `json:"syncRequestTimeout"`              // default: 0 == infinite, which means SyncRequestChannel's need to be closed manually by the application or else there will be a memory leak
	TcpTimeoutMs                    uint64 `json:"tcpTimeoutMs"`                    // default: 0 == block forever
	MaxConnectionAttempts           uint64 `json:"maxConnectionAttempts"`           // default: 0 == infinite
	ConnectionAttemptDelayMs        uint64 `json:"connectionAttemptDelay"`          // default: 0 (delay after failed connection attempt)
	StopAfterOutgoingConnectionLoss bool   `json:"stopAfterOutgoingConnectionLoss"` // default: false (relevant if maxConnectionAttempts is set)

	EndpointConfigs []*TcpEndpoint `json:"endpointConfigs"` // *required* (endpoint to other node's servers) (on startup, this node will attempt to establish connection to these endpoints)

	RateLimiterBytes    *TokenBucketRateLimiter `json:"outgoingConnectionRateLimiterBytes"` // *optional* (rate limiter for outgoing connections)
	RateLimiterMessages *TokenBucketRateLimiter `json:"outgoingConnectionRateLimiterMsgs"`  // *optional* (rate limiter for outgoing connections)

	TcpBufferBytes           uint32 `json:"tcpBufferBytes"`           // default: 0 == default (4KB)
	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 == unlimited (connections that attempt to send messages larger than this will be disconnected)
	MaxPayloadSize           int    `json:"maxPayloadSize"`           // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxTopicSize             int    `json:"maxTopicSize"`             // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxSyncTokenSize         int    `json:"maxSyncTokenSize"`         // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxNodeNameSize          uint64 `json:"maxNodeNameSize"`          // default: 0 == unlimited (connections that attempt to send a node name larger than this will be rejected)
}

func UnmarshalSystemgeClient(data string) *SystemgeClient {
	var systemge SystemgeClient
	json.Unmarshal([]byte(data), &systemge)
	return &systemge
}

type SystemgeServer struct {
	Name string `json:"name"` // *required*

	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*

	TcpTimeoutMs                                uint64 `json:"tcpTimeoutMs"`                                // default: 0 == block forever
	ProcessMessagesOfEachConnectionSequentially bool   `json:"processMessagesOfEachConnectionSequentially"` // default: false (if true, the server will handle messages from the same incoming connection sequentially) (if >1 incomming connection, multiple message handlers may run concurrently)
	ProcessAllMessagesSequentially              bool   `json:"processAllMessagesSequentially"`              // default: false (overrides ProcessMessagesOfEachConnectionSequentially) (guarantees, that only one message handler runs at a time)
	ProcessAllMessagesSequentiallyChannelSize   int    `json:"processAllMessagesSequentiallyChannelSize"`   // default: 0 == no guarantee on order of arrival (if >0, the order of arrival is guaranteed as long as the channel is never full (does NOT guarantee that messages arrive in the same chronological order as they were sent (technically not possible with multiple incoming connections and without a global clock)))

	ServerConfig *TcpServer   `json:"serverConfig"` // *required* (the configuration of this node's server)
	Endpoint     *TcpEndpoint `json:"endpoint"`     // *optional* (the configuration of this node's endpoint) (can be shared with other nodes to let them connect during runtime)

	RateLimterBytes     *TokenBucketRateLimiter `json:"incomingConnectionRateLimiterBytes"` // *optional* (rate limiter for incoming connections)
	RateLimiterMessages *TokenBucketRateLimiter `json:"incomingConnectionRateLimiterMsgs"`  // *optional* (rate limiter for incoming connections)

	IpRateLimiter *IpRateLimiter `json:"ipRateLimiter"` // *optional* (rate limiter for incoming connections) (allows to limit the number of incoming connection attempts from the same IP) (it is more efficient to use a firewall for this purpose)

	TcpBufferBytes           uint32 `json:"tcpBufferBytes"`           // default: 0 == default (4KB)
	IncomingMessageByteLimit uint64 `json:"incomingMessageByteLimit"` // default: 0 == unlimited (connections that attempt to send messages larger than this will be disconnected)
	MaxPayloadSize           int    `json:"maxPayloadSize"`           // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxTopicSize             int    `json:"maxTopicSize"`             // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxSyncTokenSize         int    `json:"maxSyncTokenSize"`         // default: <=0 == unlimited (messages that exceed this limit will be skipped)
	MaxNodeNameSize          uint64 `json:"maxNodeNameSize"`          // default: 0 == unlimited (connections that attempt to send a node name larger than this will be rejected)
}

func UnmarshalSystemgeServer(data string) *SystemgeServer {
	var systemge SystemgeServer
	json.Unmarshal([]byte(data), &systemge)
	return &systemge
}
