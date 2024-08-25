package Config

import "encoding/json"

type MessageBrokerServer struct {
	SystemgeServerConfig  *SystemgeServer  `json:"systemgeServerConfig"` // *required*
	DashboardClientConfig *DashboardClient `json:"dashboardConfig"`      // *required*

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

	AsyncTopics []string `json:"asyncTopics"` // *required*
	SyncTopics  []string `json:"syncTopics"`  // *required*

	ConnectionConfig *SystemgeConnection `json:"connectionConfig"` // *required*
	EndpointConfig   *TcpEndpoint        `json:"endpointConfig"`   // *required*
}

func UnmarshalMessageBrokerClient(data string) *MessageBrokerClient {
	var messageBroker MessageBrokerClient
	json.Unmarshal([]byte(data), &messageBroker)
	return &messageBroker
}
