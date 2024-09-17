package DashboardClientSystemgeConnection

import (
	"sync"

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
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
)

func New(name string, config *Config.DashboardClient, systemgeConnection SystemgeConnection.SystemgeConnection, messageHandler SystemgeMessageHandler.MessageHandler, getMetricsFunc func() map[string]*Metrics.Metrics, commands Commands.Handlers) *DashboardClient.Client {
	if systemgeConnection == nil {
		panic("customService is nil")
	}
	var metrics map[string]*Metrics.Metrics
	if getMetricsFunc != nil {
		metrics = getMetricsFunc()
	} else {
		metrics = map[string]*Metrics.Metrics{}
	}
	mutex := sync.Mutex{}
	var processingLoopStopChannel chan<- bool
	Metrics.Merge(metrics, systemgeConnection.GetMetrics())
	return DashboardClient.New(
		name,
		config,
		SystemgeMessageHandler.NewTopicExclusiveMessageHandler(
			nil,
			SystemgeMessageHandler.SyncMessageHandlers{
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
					var metrics map[string]*Metrics.Metrics
					if getMetricsFunc != nil {
						metrics = getMetricsFunc()
					} else {
						metrics = map[string]*Metrics.Metrics{}
					}
					Metrics.Merge(metrics, systemgeConnection.GetMetrics())
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
					mutex.Lock()
					defer mutex.Unlock()
					if processingLoopStopChannel != nil {
						return "", Error.New("Processing loop is already running", nil)
					}
					stopChannel, _ := SystemgeMessageHandler.StartMessageHandlingLoop_Sequentially(systemgeConnection, messageHandler)
					processingLoopStopChannel = stopChannel
					return "", nil
				},
				DashboardHelpers.TOPIC_START_PROCESSINGLOOP_CONCURRENTLY: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					mutex.Lock()
					defer mutex.Unlock()
					if processingLoopStopChannel != nil {
						return "", Error.New("Processing loop is already running", nil)
					}
					stopChannel, _ := SystemgeMessageHandler.StartMessageHandlingLoop_Concurrently(systemgeConnection, messageHandler)
					processingLoopStopChannel = stopChannel
					return "", nil
				},
				DashboardHelpers.TOPIC_STOP_PROCESSINGLOOP: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					mutex.Lock()
					defer mutex.Unlock()
					if processingLoopStopChannel == nil {
						return "", Error.New("Processing loop is not running", nil)
					}
					close(processingLoopStopChannel)
					processingLoopStopChannel = nil
					return "", nil
				},
				DashboardHelpers.TOPIC_IS_PROCESSING_LOOP_RUNNING: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					return Helpers.BoolToString(processingLoopStopChannel != nil), nil
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
					processingLoopStopChannel != nil,
					systemgeConnection.AvailableMessageCount(),
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
