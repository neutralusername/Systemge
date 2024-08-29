package MessageBroker

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Dashboard"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func NewMessageBrokerClient(config *Config.MessageBrokerClient, systemgeMessageHandler SystemgeConnection.MessageHandler, dashboardCommands Commands.Handlers) (*SystemgeConnection.SystemgeConnection, error) {
	if config.ConnectionConfig == nil {
		return nil, Error.New("ConnectionConfig is required", nil)
	}
	if config.EndpointConfig == nil {
		return nil, Error.New("EndpointConfig is required", nil)
	}
	if config.ConnectionConfig.TcpBufferBytes == 0 {
		config.ConnectionConfig.TcpBufferBytes = 1024 * 4
	}
	systemgeConnection, err := SystemgeConnection.EstablishConnection(config.ConnectionConfig, config.EndpointConfig, config.Name, config.MaxServerNameLength)
	if err != nil {
		return nil, Error.New("Failed to establish connection", err)
	}
	if _, err := systemgeConnection.SyncRequestBlocking(Message.TOPIC_ADD_ASYNC_TOPICS, Helpers.JsonMarshal(systemgeMessageHandler.GetAsyncTopics())); err != nil {
		systemgeConnection.Close()
		return nil, Error.New("Failed to add async topics", err)
	}
	if _, err := systemgeConnection.SyncRequestBlocking(Message.TOPIC_ADD_SYNC_TOPICS, Helpers.JsonMarshal(systemgeMessageHandler.GetSyncTopics())); err != nil {
		systemgeConnection.Close()
		return nil, Error.New("Failed to add sync topics", err)
	}
	dashboardClient := Dashboard.NewClient(config.DashboardClientConfig, nil, systemgeConnection.Close, systemgeConnection.GetMetrics, systemgeConnection.IsClosed, dashboardCommands)
	if err := dashboardClient.Start(); err != nil {
		systemgeConnection.Close()
		return nil, Error.New("Failed to start dashboard client", err)
	}
	go func() {
		<-systemgeConnection.GetCloseChannel()
		dashboardClient.Stop()
	}()
	if err := systemgeConnection.StartProcessingLoopSequentially(systemgeMessageHandler); err != nil {
		systemgeConnection.Close()
		return nil, Error.New("Failed to start processing loop", err)
	}
	return systemgeConnection, nil
}
