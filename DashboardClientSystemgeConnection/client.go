package DashboardClientSystemgeConnection

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
)

func New(name string, config *Config.DashboardClient, systemgeConnection SystemgeConnection.SystemgeConnection, messageHandler SystemgeConnection.TopicHandler, getMetricsFunc func() map[string]*Metrics.Metrics, commands Commands.Handlers) *DashboardClient.Client {
	if systemgeConnection == nil {
		panic("customService is nil")
	}

	metricsTypes := Metrics.NewMetricsTypes()
	if getMetricsFunc != nil {
		metricsTypes.Merge(getMetricsFunc())
	}
	metricsTypes.Merge(systemgeConnection.GetMetrics())

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
					return Helpers.IntToString(systemgeConnection.GetStatus()), nil
				},
				DashboardHelpers.TOPIC_GET_METRICS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					metricsTypes := Metrics.NewMetricsTypes()
					if getMetricsFunc != nil {
						metricsTypes.Merge(getMetricsFunc())
					}
					metricsTypes.Merge(systemgeConnection.GetMetrics())
					return Helpers.JsonMarshal(metricsTypes), nil
				},
				DashboardHelpers.TOPIC_STOP: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					err := systemgeConnection.Close()
					if err != nil {
						return "", err
					}
					return Helpers.IntToString(Status.STOPPED), nil
				},
				DashboardHelpers.TOPIC_START_PROCESSINGLOOP_SEQUENTIALLY: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					err := systemgeConnection.StartMessageHandlingLoop_Sequentially(messageHandler)
					if err != nil {
						return "", Error.New("Failed to start processing loop", err)
					}
					return "", nil
				},
				DashboardHelpers.TOPIC_START_PROCESSINGLOOP_CONCURRENTLY: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					err := systemgeConnection.StartMessageHandlingLoop_Concurrently(messageHandler)
					if err != nil {
						return "", Error.New("Failed to start processing loop", err)
					}
					return "", nil
				},
				DashboardHelpers.TOPIC_STOP_PROCESSINGLOOP: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					err := systemgeConnection.StopMessageHandlingLoop()
					if err != nil {
						return "", Error.New("Failed to stop processing loop", err)
					}
					return "", nil
				},
				DashboardHelpers.TOPIC_IS_PROCESSING_LOOP_RUNNING: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					return Helpers.BoolToString(systemgeConnection.IsMessageHandlingLoopStarted()), nil
				},
				DashboardHelpers.TOPIC_PROCESS_NEXT_MESSAGE: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
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
				DashboardHelpers.TOPIC_UNPROCESSED_MESSAGE_COUNT: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					return Helpers.Uint32ToString(systemgeConnection.AvailableMessageCount()), nil
				},
				DashboardHelpers.TOPIC_SYNC_REQUEST: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					message, err := Message.Deserialize([]byte(message.GetPayload()), connection.GetName())
					if err != nil {
						return "", Error.New("Failed to deserialize message", err)
					}
					response, err := systemgeConnection.SyncRequestBlocking(message.GetTopic(), message.GetPayload())
					if err != nil {
						return "", Error.New("Failed to complete sync request", err)
					}
					return string(response.Serialize()), nil
				},
				DashboardHelpers.TOPIC_ASYNC_MESSAGE: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					message, err := Message.Deserialize([]byte(message.GetPayload()), connection.GetName())
					if err != nil {
						return "", Error.New("Failed to deserialize message", err)
					}
					err = systemgeConnection.AsyncMessage(message.GetTopic(), message.GetPayload())
					if err != nil {
						return "", Error.New("Failed to handle async message", err)
					}
					return "", nil
				},
			},
			nil, nil,
			1000,
		),
		func() (string, error) {
			pageMarshalled, err := DashboardHelpers.NewPage(
				DashboardHelpers.NewSystemgeConnectionClient(
					name,
					commands.GetKeyBoolMap(),
					systemgeConnection.GetStatus(),
					systemgeConnection.IsMessageHandlingLoopStarted(),
					systemgeConnection.AvailableMessageCount(),
					DashboardHelpers.NewDashboardMetrics(metricsTypes),
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
