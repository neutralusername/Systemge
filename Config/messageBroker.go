package Config

import "encoding/json"

type MessageBrokerServer struct {
	SystemgeServerConfig *SystemgeServer `json:"systemgeServerConfig"` // *required*

	AsyncTopics []string `json:"asyncTopics"`
	SyncTopics  []string `json:"syncTopics"`

	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*
	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
}

func UnmarshalMessageBrokerServer(data string) *MessageBrokerServer {
	var messageBroker MessageBrokerServer
	json.Unmarshal([]byte(data), &messageBroker)
	return &messageBroker
}

type MessageBrokerClient struct {
	ServerTcpSystemgeConnectionConfig   *TcpSystemgeConnection `json:"serverTcpSystemgeConnectionConfig"`   // *required*
	ResolverTcpSystemgeConnectionConfig *TcpSystemgeConnection `json:"resolverTcpSystemgeConnectionConfig"` // *required*

	ResolverTcpClientConfigs []*TcpClient `json:"resolverTcpClientConfigs"`

	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*
	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*

	TopicResolutionLifetimeMs        uint64 `json:"topicResolutionLifetimeMs"`     // default: 0 == until disconnect
	MaxServerNameLength              int    `json:"maxServerNameLength"`           // default: <=0 == no limit
	ResolutionAttemptRetryIntervalMs uint32 `json:"subscribedTopicsRetryMs"`       // default: 0 == no delay
	ResolutionAttemptMaxAttempts     uint32 `json:"subscribedTopicsRetryAttempts"` // default: 0 == no limit

	AsyncTopics []string `json:"asyncTopics"`
	SyncTopics  []string `json:"syncTopics"`
}

func UnmarshalMessageBrokerClient(data string) *MessageBrokerClient {
	var messageBroker MessageBrokerClient
	json.Unmarshal([]byte(data), &messageBroker)
	return &messageBroker
}

type MessageBrokerResolver struct {
	SystemgeServerConfig *SystemgeServer `json:"systemgeServerConfig"` // *required*

	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*
	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*

	MaxClientNameLength int `json:"maxClientNameLength"` // default: <=0 == no limit

	AsyncTopicClientConfigs map[string]*TcpClient `json:"asyncTopicClientConfigs"`
	SyncTopicClientConfigs  map[string]*TcpClient `json:"syncTopicClientConfigs"`
}
