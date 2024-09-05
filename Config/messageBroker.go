package Config

import "encoding/json"

type MessageBrokerServer struct {
	SystemgeServerConfig  *SystemgeServer  `json:"systemgeServerConfig"`  // *required*
	DashboardClientConfig *DashboardClient `json:"dashboardClientConfig"` // *required*

	AsyncTopics []string `json:"asyncTopics"` // *required*
	SyncTopics  []string `json:"syncTopics"`  // *required*

	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *required*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *required*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *required*
	MailerConfig      *Mailer `json:"mailerConfig"`      // *required*
}

func UnmarshalMessageBrokerServer(data string) *MessageBrokerServer {
	var messageBroker MessageBrokerServer
	json.Unmarshal([]byte(data), &messageBroker)
	return &messageBroker
}

type MessageBrokerClient struct {
	ConnectionConfig         *TcpConnection `json:"outConnectionConfig"`      // *required*
	ResolverConnectionConfig *TcpConnection `json:"resolverConnectionConfig"` // *required*

	DashboardClientConfig *DashboardClient `json:"dashboardClientConfig"` // *optional* (it is a valid choice to leave this nil and initialize your own DashboardClient)

	ResolverClientConfigs []*TcpClient `json:"resolverClientConfigs"`

	TopicResolutionLifetimeMs uint64 `json:"outTopicResolutionLifetimeMs"` // default: 0 == until disconnect

	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*
	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*

	MaxServerNameLength int `json:"maxServerNameLength"` // *optional*

	AsyncTopics []string `json:"asyncTopics"`
	SyncTopics  []string `json:"syncTopics"`
}

func UnmarshalMessageBrokerClient(data string) *MessageBrokerClient {
	var messageBroker MessageBrokerClient
	json.Unmarshal([]byte(data), &messageBroker)
	return &messageBroker
}

type MessageBrokerResolver struct {
	SystemgeServerConfig  *SystemgeServer  `json:"systemgeServerConfig"`  // *required*
	DashboardClientConfig *DashboardClient `json:"dashboardClientConfig"` // *required*

	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *required*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *required*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *required*
	MailerConfig      *Mailer `json:"mailerConfig"`      // *required*

	MaxClientNameLength int `json:"maxClientNameLength"` // *required*

	AsyncTopicClientConfigs map[string]*TcpClient `json:"asyncTopicClientConfigs"` // *required*
	SyncTopicClientConfigs  map[string]*TcpClient `json:"syncTopicClientConfigs"`  // *required*
}
