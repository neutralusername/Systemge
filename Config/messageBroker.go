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
	Name string `json:"name"` // *required*

	InConnectionConfig  *SystemgeConnection `json:"inConnectionConfig"`  // *required*
	OutConnectionConfig *SystemgeConnection `json:"outConnectionConfig"` // *required*

	ResolverConnectionConfig *SystemgeConnection `json:"resolverConnectionConfig"` // *required*
	ResolverEndpoints        []*TcpEndpoint      `json:"resolverEndpoints"`        // *required*
	DashboardClientConfig    *DashboardClient    `json:"dashboardClientConfig"`    // *required*

	OutTopicResolutionLifetimeMs uint64 `json:"outTopicResolutionLifetimeMs"` // default: 0 == until disconnect
	InTopicResolutionLifetimeMs  uint64 `json:"inTopicResolutionLifetimeMs"`  // default: 0 == until disconnect

	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *required*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *required*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *required*
	MailerConfig      *Mailer `json:"mailerConfig"`      // *required*

	MaxServerNameLength int `json:"maxServerNameLength"` // *required*

	AsyncTopics []string `json:"asyncTopics"` // *required*
	SyncTopics  []string `json:"syncTopics"`  // *required*

}

func UnmarshalMessageBrokerClient(data string) *MessageBrokerClient {
	var messageBroker MessageBrokerClient
	json.Unmarshal([]byte(data), &messageBroker)
	return &messageBroker
}

type MessageBrokerResolver struct {
	Name string `json:"name"` // *required*

	SystemgeServerConfig  *SystemgeServer  `json:"systemgeServerConfig"`  // *required*
	DashboardClientConfig *DashboardClient `json:"dashboardClientConfig"` // *required*

	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *required*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *required*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *required*
	MailerConfig      *Mailer `json:"mailerConfig"`      // *required*

	MaxClientNameLength int `json:"maxClientNameLength"` // *required*

	AsyncTopicEndpoints map[string]*TcpEndpoint `json:"asyncTopicResolutions"` // *required*
	SyncTopicEndpoints  map[string]*TcpEndpoint `json:"syncTopicResolutions"`  // *required*
}
