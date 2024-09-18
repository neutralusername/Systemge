package DashboardClientSystemgeServer

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/DashboardClient"
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Metrics"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
)

func New(name string, config *Config.DashboardClient, systemgeServer *SystemgeServer.SystemgeServer, messageHandler SystemgeConnection.MessageHandler, getMetricsFunc func() map[string]*Metrics.Metrics, commands Commands.Handlers) *DashboardClient.Client {
	if systemgeServer == nil {
		panic("customService is nil")
	}
	metricsTypes := Metrics.NewMetricsTypes()
	if getMetricsFunc != nil {
		metricsTypes.Merge(getMetricsFunc())
	}
	metricsTypes.Merge(systemgeServer.GetMetrics())

	systemgeConnectionChildren := map[string]*DashboardHelpers.SystemgeConnectionChild{}
	for _, systemgeConnection := range systemgeServer.GetConnections() {
		systemgeConnectionChildren[systemgeConnection.GetName()] = DashboardHelpers.NewSystemgeConnectionChild(systemgeConnection.GetName(), systemgeConnection.IsMessageHandlingLoopStarted(), systemgeConnection.AvailableMessageCount())
	}

	return DashboardClient.New(
		name,
		config,
		SystemgeConnection.NewTopicExclusiveMessageHandler(
			nil,
			SystemgeConnection.SyncMessageHandlers{
				DashboardHelpers.TOPIC_COMMAND: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					command, err := DashboardHelpers.UnmarshalCommand(message.GetPayload())
					if err != nil {
						return "", err
					}
					return commands.Execute(command.Command, command.Args)
				},
				DashboardHelpers.TOPIC_GET_STATUS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					return Helpers.IntToString(systemgeServer.GetStatus()), nil
				},
				DashboardHelpers.TOPIC_GET_METRICS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					metricsTypes := Metrics.NewMetricsTypes()
					if getMetricsFunc != nil {
						metricsTypes.Merge(getMetricsFunc())
					}
					metricsTypes.Merge(systemgeServer.GetMetrics())
					return Helpers.JsonMarshal(metricsTypes), nil
				},
				DashboardHelpers.TOPIC_STOP: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					err := systemgeServer.Stop()
					if err != nil {
						return "", err
					}
					return Helpers.IntToString(Status.STOPPED), nil
				},
				DashboardHelpers.TOPIC_START: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					err := systemgeServer.Start()
					if err != nil {
						return "", err
					}
					return Helpers.IntToString(Status.STARTED), nil
				},
				DashboardHelpers.TOPIC_SYNC_REQUEST: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					message, err := Message.Deserialize([]byte(message.GetPayload()), connection.GetName())
					if err != nil {
						return "", Error.New("Failed to deserialize message", err)
					}
					responses, err := systemgeServer.SyncRequestBlocking(message.GetTopic(), message.GetPayload())
					if err != nil {
						return "", Error.New("Failed to complete sync request", err)
					}
					return string(Helpers.JsonMarshal(responses)), nil
				},
				DashboardHelpers.TOPIC_ASYNC_MESSAGE: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					message, err := Message.Deserialize([]byte(message.GetPayload()), connection.GetName())
					if err != nil {
						return "", Error.New("Failed to deserialize message", err)
					}
					err = systemgeServer.AsyncMessage(message.GetTopic(), message.GetPayload())
					if err != nil {
						return "", Error.New("Failed to handle async message", err)
					}
					return "", nil
				},
				DashboardHelpers.TOPIC_GET_CONNECTIONS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					systemgeConnectionChildren := map[string]*DashboardHelpers.SystemgeConnectionChild{}
					for _, systemgeConnection := range systemgeServer.GetConnections() {
						systemgeConnectionChildren[systemgeConnection.GetName()] = DashboardHelpers.NewSystemgeConnectionChild(systemgeConnection.GetName(), systemgeConnection.IsMessageHandlingLoopStarted(), systemgeConnection.AvailableMessageCount())
					}
					return Helpers.JsonMarshal(systemgeConnectionChildren), nil
				},

				DashboardHelpers.TOPIC_START_PROCESSINGLOOP_SEQUENTIALLY: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					systemgeConnection := systemgeServer.GetConnection(message.GetPayload())
					if systemgeConnection == nil {
						return "", Error.New("Connection not found", nil)
					}
					err := systemgeConnection.StartMessageHandlingLoop_Sequentially(messageHandler)
					if err != nil {
						return "", err
					}
					return "", nil
				},
				DashboardHelpers.TOPIC_START_PROCESSINGLOOP_CONCURRENTLY: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					systemgeConnection := systemgeServer.GetConnection(message.GetPayload())
					if systemgeConnection == nil {
						return "", Error.New("Connection not found", nil)
					}
					err := systemgeConnection.StartMessageHandlingLoop_Concurrently(messageHandler)
					if err != nil {
						return "", err
					}
					return "", nil
				},
				DashboardHelpers.TOPIC_STOP_PROCESSINGLOOP: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					systemgeConnection := systemgeServer.GetConnection(message.GetPayload())
					if systemgeConnection == nil {
						return "", Error.New("Connection not found", nil)
					}
					err := systemgeConnection.StopMessageHandlingLoop()
					if err != nil {
						return "", err
					}
					return "", nil
				},
				DashboardHelpers.TOPIC_PROCESS_NEXT_MESSAGE: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					systemgeConnection := systemgeServer.GetConnection(message.GetPayload())
					if systemgeConnection == nil {
						return "", Error.New("Connection not found", nil)
					}
					message, err := systemgeConnection.GetNextMessage()
					if err != nil {
						return "", Error.New("Failed to get next message", err)
					}
					if message.GetSyncToken() == "" {
						err := messageHandler.HandleAsyncMessage(systemgeConnection, message)
						if err != nil {
							return "", Error.New("Failed to handle async message with topic \""+message.GetTopic()+"\" and payload \""+message.GetPayload()+"\"", err)
						}
						return "Handled async message with topic \"" + message.GetTopic() + "\" and payload \"" + message.GetPayload() + "\"", nil
					}
					if responsePayload, err := messageHandler.HandleSyncRequest(systemgeConnection, message); err != nil {
						if err := systemgeConnection.SyncResponse(message, false, err.Error()); err != nil {
							return "", Error.New("Failed to handle sync request with topic \""+message.GetTopic()+"\" and payload \""+message.GetPayload()+"\" and failed to send failure response \""+err.Error()+"\"", err)
						}
						return "Failed to handle sync request with topic \"" + message.GetTopic() + "\" and payload \"" + message.GetPayload() + "\" and sent failure response \"" + err.Error() + "\"", nil
					} else {
						if err := systemgeConnection.SyncResponse(message, true, responsePayload); err != nil {
							return "", Error.New("Handled sync request with topic \""+message.GetTopic()+"\" and payload \""+message.GetPayload()+"\" and failed to send success response \""+responsePayload+"\"", err)
						}
						return "Handled sync request with topic \"" + message.GetTopic() + "\" and payload \"" + message.GetPayload() + "\" and sent success response \"" + responsePayload + "\"", nil
					}
				},
			},
			nil, nil,
			1000,
		),
		func() (string, error) {
			pageMarshalled, err := DashboardHelpers.NewPage(
				DashboardHelpers.NewSystemgeServerClient(
					name,
					commands.GetKeyBoolMap(),
					systemgeServer.GetStatus(),
					DashboardHelpers.NewDashboardMetrics(metricsTypes),
					systemgeConnectionChildren,
				),
				DashboardHelpers.CLIENT_TYPE_SYSTEMGECONNECTION,
			).Marshal()
			if err != nil {
				panic(err)
			}
			return string(pageMarshalled), nil
		},
	)
}
