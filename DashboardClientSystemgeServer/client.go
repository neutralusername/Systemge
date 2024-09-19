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

// frontend not implemented nor is this tested (use DashboardClientCustomService for now)
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
				DashboardHelpers.TOPIC_GET_SYSTEMGE_CONNECTION_CHILDREN: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					systemgeConnectionChildren := map[string]*DashboardHelpers.SystemgeConnectionChild{}
					for _, systemgeConnection := range systemgeServer.GetConnections() {
						systemgeConnectionChildren[systemgeConnection.GetName()] = DashboardHelpers.NewSystemgeConnectionChild(systemgeConnection.GetName(), systemgeConnection.IsMessageHandlingLoopStarted(), systemgeConnection.AvailableMessageCount())
					}
					return Helpers.JsonMarshal(systemgeConnectionChildren), nil
				},

				DashboardHelpers.TOPIC_COMMAND: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					command, err := DashboardHelpers.UnmarshalCommand(message.GetPayload())
					if err != nil {
						return "", err
					}
					return commands.Execute(command.Command, command.Args)
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

				DashboardHelpers.TOPIC_MULTI_SYNC_REQUEST: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					messageWithRecipients, err := DashboardHelpers.UnmarshalMultiMessage([]byte(message.GetPayload()))
					if err != nil {
						return "", Error.New("Failed to deserialize message", err)
					}
					responses, err := systemgeServer.SyncRequestBlocking(messageWithRecipients.Message.GetTopic(), messageWithRecipients.Message.GetPayload(), messageWithRecipients.Recipients...)
					if err != nil {
						return "", Error.New("Failed to complete sync request", err)
					}
					return string(Helpers.JsonMarshal(responses)), nil
				},
				DashboardHelpers.TOPIC_MULTI_ASYNC_MESSAGE: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					messageWithRecipients, err := DashboardHelpers.UnmarshalMultiMessage([]byte(message.GetPayload()))
					if err != nil {
						return "", Error.New("Failed to deserialize message", err)
					}
					err = systemgeServer.AsyncMessage(messageWithRecipients.Message.GetTopic(), messageWithRecipients.Message.GetPayload(), messageWithRecipients.Recipients...)
					if err != nil {
						return "", Error.New("Failed to handle async message", err)
					}
					return "", nil
				},

				DashboardHelpers.TOPIC_CLOSE_CHILD: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					systemgeConnection := systemgeServer.GetConnection(message.GetPayload())
					if systemgeConnection == nil {
						return "", Error.New("Connection not found", nil)
					}
					err := systemgeConnection.Close()
					if err != nil {
						return "", err
					}
					return Helpers.IntToString(Status.STOPPED), nil
				},
				DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_SEQUENTIALLY_CHILD: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
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
				DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_CONCURRENTLY_CHILD: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
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
				DashboardHelpers.TOPIC_STOP_MESSAGE_HANDLING_LOOP_CHILD: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
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
				DashboardHelpers.TOPIC_HANDLE_NEXT_MESSAGE_CHILD: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					systemgeConnection := systemgeServer.GetConnection(message.GetPayload())
					if systemgeConnection == nil {
						return "", Error.New("Connection not found", nil)
					}
					message, err := systemgeConnection.GetNextMessage()
					if err != nil {
						return "", Error.New("Failed to get next message", err)
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
								handleNextMessageResult.Error = Error.New(handleNextMessageResult.Error, err).Error()
							}
						} else {
							handleNextMessageResult.HandlingSucceeded = true
							handleNextMessageResult.ResultPayload = responsePayload
							if err := systemgeConnection.SyncResponse(message, true, responsePayload); err != nil {
								handleNextMessageResult.Error = err.Error()
							}
						}
					}
					return handleNextMessageResult.Marshal(), nil
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
