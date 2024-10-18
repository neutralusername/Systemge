package configs

import "encoding/json"

type MessageBrokerClient struct {
	ServerTcpSystemgeConnectionConfig   *TcpConnection `json:"serverTcpSystemgeConnectionConfig"`   // *required*
	ResolverTcpSystemgeConnectionConfig *TcpConnection `json:"resolverTcpSystemgeConnectionConfig"` // *required*

	ResolverTcpClientConfigs []*TcpClient `json:"resolverTcpClientConfigs"`

	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*
	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*

	TopicResolutionLifetimeMs        uint64 `json:"topicResolutionLifetimeMs"`        // default: 0 == until disconnect
	MaxServerNameLength              int    `json:"maxServerNameLength"`              // default: <=0 == no limit
	ResolutionAttemptRetryIntervalMs uint32 `json:"resolutionAttemptRetryIntervalMs"` // default: 0 == no delay
	ResolutionMaxAttempts            uint32 `json:"resolutionMaxAttempts"`            // default: 0 == no limit

	AsyncTopics []string `json:"asyncTopics"`
	SyncTopics  []string `json:"syncTopics"`
}

func UnmarshalMessageBrokerClient(data string) *MessageBrokerClient {
	var messageBroker MessageBrokerClient
	json.Unmarshal([]byte(data), &messageBroker)
	return &messageBroker
}
