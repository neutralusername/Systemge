package DashboardClientSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/DashboardClient"
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func New(name string, config *Config.DashboardClient, systemgeConnection SystemgeConnection.SystemgeConnection, messageHandler SystemgeConnection.MessageHandler, getMetricsFunc func() map[string]map[string]uint64, commands Commands.Handlers) *DashboardClient.Client {
	if systemgeConnection == nil {
		panic("customService is nil")
	}
	var metrics map[string]map[string]uint64 = make(map[string]map[string]uint64)
	if getMetricsFunc != nil {
		metrics = getMetricsFunc()
	}
	DashboardHelpers.MergeMetrics(metrics, systemgeConnection.GetMetrics())
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
					var metrics map[string]map[string]uint64 = make(map[string]map[string]uint64)
					if getMetricsFunc != nil {
						metrics = getMetricsFunc()
					}
					DashboardHelpers.MergeMetrics(metrics, systemgeConnection.GetMetrics())
					return Helpers.JsonMarshal(DashboardHelpers.NewDashboardMetrics(metrics)), nil
				},
				DashboardHelpers.TOPIC_STOP: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					err := systemgeConnection.Close()
					if err != nil {
						return "", err
					}
					return Helpers.IntToString(Status.STOPPED), nil
				},
				DashboardHelpers.TOPIC_START_PROCESSINGLOOP_SEQUENTIALLY: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					err := systemgeConnection.StartProcessingLoopSequentially(messageHandler)
					if err != nil {
						return "", err
					}
					return "", nil
				},
				DashboardHelpers.TOPIC_START_PROCESSINGLOOP_CONCURRENTLY: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					err := systemgeConnection.StartProcessingLoopConcurrently(messageHandler)
					if err != nil {
						return "", err
					}
					return "", nil
				},
				DashboardHelpers.TOPIC_STOP_PROCESSINGLOOP: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					err := systemgeConnection.StopProcessingLoop()
					if err != nil {
						return "", err
					}
					return "", nil
				},
				DashboardHelpers.TOPIC_IS_PROCESSING_LOOP_RUNNING: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					return Helpers.BoolToString(systemgeConnection.IsProcessingLoopRunning()), nil
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
					return Helpers.Uint32ToString(systemgeConnection.UnprocessedMessagesCount()), nil
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
					systemgeConnection.UnprocessedMessagesCount(),
					DashboardHelpers.NewDashboardMetrics(metrics),
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
