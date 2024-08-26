package MessageBroker

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Dashboard"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func NewMessageBrokerClient(config *Config.MessageBrokerClient) (*SystemgeConnection.SystemgeConnection, error) {
	if config.ConnectionConfig == nil {
		return nil, Error.New("ConnectionConfig is required", nil)
	}
	if config.EndpointConfig == nil {
		return nil, Error.New("EndpointConfig is required", nil)
	}
	systemgeConnection, err := SystemgeConnection.EstablishConnection(config.ConnectionConfig, config.EndpointConfig, config.Name, config.MaxServerNameLength)
	if err != nil {
		return nil, Error.New("Failed to establish connection", err)
	}
	if _, err := systemgeConnection.SyncRequestBlocking(Message.TOPIC_ADD_ASYNC_TOPICS, Helpers.JsonMarshal(config.AsyncTopics)); err != nil {
		systemgeConnection.Close()
		return nil, Error.New("Failed to add async topics", err)
	}
	if _, err := systemgeConnection.SyncRequestBlocking(Message.TOPIC_ADD_SYNC_TOPICS, Helpers.JsonMarshal(config.SyncTopics)); err != nil {
		systemgeConnection.Close()
		return nil, Error.New("Failed to add sync topics", err)
	}
	if err := Dashboard.NewClient(config.DashboardClientConfig, nil, nil, nil, nil, nil).Start(); err != nil {
		systemgeConnection.Close()
		return nil, Error.New("Failed to start dashboard client", err)
	}
	return systemgeConnection, nil
}
