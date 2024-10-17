package DashboardClientSystemgeConnection

import (
	"github.com/neutralusername/systemge/Commands"
	"github.com/neutralusername/systemge/Config"
	"github.com/neutralusername/systemge/DashboardClient"
	"github.com/neutralusername/systemge/DashboardHelpers"
	"github.com/neutralusername/systemge/Event"
	"github.com/neutralusername/systemge/Message"
	"github.com/neutralusername/systemge/Metrics"
	"github.com/neutralusername/systemge/SystemgeConnection"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/status"
)

// frontend not implemented nor is this tested (use DashboardClientCustomService for now)
func New(name string, config *Config.DashboardClient, systemgeConnection SystemgeConnection.SystemgeConnection, messageHandler SystemgeConnection.MessageHandler, getMetricsFunc func() map[string]*Metrics.Metrics, commands Commands.Handlers, eventHandler Event.Handler) (*DashboardClient.Client, error) {
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
		nil,
		SystemgeConnection.SyncMessageHandlers{
			DashboardHelpers.TOPIC_GET_STATUS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				return helpers.IntToString(systemgeConnection.GetStatus()), nil
			},
			DashboardHelpers.TOPIC_GET_METRICS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				metricsTypes := Metrics.NewMetricsTypes()
				if getMetricsFunc != nil {
					metricsTypes.Merge(getMetricsFunc())
				}
				metricsTypes.Merge(systemgeConnection.GetMetrics())
				return helpers.JsonMarshal(metricsTypes), nil
			},
			DashboardHelpers.TOPIC_IS_MESSAGE_HANDLING_LOOP_STARTED: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				return helpers.BoolToString(systemgeConnection.IsMessageHandlingLoopStarted()), nil
			},
			DashboardHelpers.TOPIC_UNHANDLED_MESSAGE_COUNT: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				return helpers.Uint32ToString(systemgeConnection.AvailableMessageCount()), nil
			},

			DashboardHelpers.TOPIC_COMMAND: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				command, err := DashboardHelpers.UnmarshalCommand(message.GetPayload())
				if err != nil {
					return "", err
				}
				return commands.Execute(command.Command, command.Args)
			},
			DashboardHelpers.TOPIC_STOP: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				err := systemgeConnection.Close()
				if err != nil {
					return "", err
				}
				return helpers.IntToString(status.Stopped), nil
			},
			DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_SEQUENTIALLY: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				err := systemgeConnection.StartMessageHandlingLoop(messageHandler, true)
				if err != nil {
					return "", err
				}
				return "", nil
			},
			DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_CONCURRENTLY: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				err := systemgeConnection.StartMessageHandlingLoop(messageHandler, false)
				if err != nil {
					return "", err
				}
				return "", nil
			},
			DashboardHelpers.TOPIC_STOP_MESSAGE_HANDLING_LOOP: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				err := systemgeConnection.StopMessageHandlingLoop()
				if err != nil {
					return "", err
				}
				return "", nil
			},
			DashboardHelpers.TOPIC_HANDLE_NEXT_MESSAGE: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				message, err := systemgeConnection.RetrieveNextMessage()
				if err != nil {
					return "", err
				}
				handleNextMessageResult := DashboardHelpers.HandleNextMessageResult{
					Message: message,
				}
				if message.GetSyncToken() == "" {
					err := messageHandler.HandleAsyncMessage(systemgeConnection, message)
					if err != nil {
						handleNextMessageResult.Error = err.Error()
						handleNextMessageResult.HandlingSucceeded = false
					} else {
						handleNextMessageResult.HandlingSucceeded = true
					}
				} else {
					if responsePayload, err := messageHandler.HandleSyncRequest(systemgeConnection, message); err != nil {
						handleNextMessageResult.Error = err.Error()
						handleNextMessageResult.HandlingSucceeded = false
						if err := systemgeConnection.SyncResponse(message, false, err.Error()); err != nil {
							handleNextMessageResult.Error = Event.New(handleNextMessageResult.Error, err).Error()
						}
					} else {
						handleNextMessageResult.HandlingSucceeded = true
						handleNextMessageResult.ResultPayload = responsePayload
						if err := systemgeConnection.SyncResponse(message, true, responsePayload); err != nil {
							handleNextMessageResult.Error = err.Error()
						}
					}
				}
				handleNextMessageResult.UnhandledMessageCount = systemgeConnection.AvailableMessageCount()
				return handleNextMessageResult.Marshal(), nil
			},
			DashboardHelpers.TOPIC_SYNC_REQUEST: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				message, err := Message.Deserialize([]byte(message.GetPayload()), connection.GetName())
				if err != nil {
					return "", err
				}
				response, err := systemgeConnection.SyncRequestBlocking(message.GetTopic(), message.GetPayload())
				if err != nil {
					return "", err
				}
				return string(response.Serialize()), nil
			},
			DashboardHelpers.TOPIC_ASYNC_MESSAGE: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				message, err := Message.Deserialize([]byte(message.GetPayload()), connection.GetName())
				if err != nil {
					return "", err
				}
				err = systemgeConnection.AsyncMessage(message.GetTopic(), message.GetPayload())
				if err != nil {
					return "", err
				}
				return "", nil
			},
		},
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
		eventHandler,
	)
}
