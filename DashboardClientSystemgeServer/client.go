package DashboardClientSystemgeServer

import (
	"errors"

	"github.com/neutralusername/systemge/Commands"
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/DashboardClient"
	"github.com/neutralusername/systemge/DashboardHelpers"
	"github.com/neutralusername/systemge/Event"
	"github.com/neutralusername/systemge/Message"
	"github.com/neutralusername/systemge/Metrics"
	"github.com/neutralusername/systemge/Server"
	"github.com/neutralusername/systemge/SystemgeConnection"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/status"
)

// frontend not implemented nor is this tested (use DashboardClientCustomService for now)
func New(name string, config *configs.DashboardClient, systemgeServer *Server.Server, messageHandler SystemgeConnection.MessageHandler, getMetricsFunc func() map[string]*Metrics.Metrics, commands Commands.Handlers, eventHandler Event.Handler) (*DashboardClient.Client, error) {
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
		nil,
		SystemgeConnection.SyncMessageHandlers{
			DashboardHelpers.TOPIC_GET_STATUS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				return helpers.IntToString(systemgeServer.GetStatus()), nil
			},
			DashboardHelpers.TOPIC_GET_METRICS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				metricsTypes := Metrics.NewMetricsTypes()
				if getMetricsFunc != nil {
					metricsTypes.Merge(getMetricsFunc())
				}
				metricsTypes.Merge(systemgeServer.GetMetrics())
				return helpers.JsonMarshal(metricsTypes), nil
			},
			DashboardHelpers.TOPIC_GET_SYSTEMGE_CONNECTION_CHILDREN: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				systemgeConnectionChildren := map[string]*DashboardHelpers.SystemgeConnectionChild{}
				for _, systemgeConnection := range systemgeServer.GetConnections() {
					systemgeConnectionChildren[systemgeConnection.GetName()] = DashboardHelpers.NewSystemgeConnectionChild(systemgeConnection.GetName(), systemgeConnection.IsMessageHandlingLoopStarted(), systemgeConnection.AvailableMessageCount())
				}
				return helpers.JsonMarshal(systemgeConnectionChildren), nil
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
				return helpers.IntToString(status.Stopped), nil
			},
			DashboardHelpers.TOPIC_START: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				err := systemgeServer.Start()
				if err != nil {
					return "", err
				}
				return helpers.IntToString(status.Started), nil
			},

			DashboardHelpers.TOPIC_MULTI_SYNC_REQUEST: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				messageWithRecipients, err := DashboardHelpers.UnmarshalMultiMessage([]byte(message.GetPayload()))
				if err != nil {
					return "", err
				}
				responses, err := systemgeServer.SyncRequestBlocking(messageWithRecipients.Message.GetTopic(), messageWithRecipients.Message.GetPayload(), messageWithRecipients.Recipients...)
				if err != nil {
					return "", err
				}
				return string(helpers.JsonMarshal(responses)), nil
			},
			DashboardHelpers.TOPIC_MULTI_ASYNC_MESSAGE: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				messageWithRecipients, err := DashboardHelpers.UnmarshalMultiMessage([]byte(message.GetPayload()))
				if err != nil {
					return "", err
				}
				err = systemgeServer.AsyncMessage(messageWithRecipients.Message.GetTopic(), messageWithRecipients.Message.GetPayload(), messageWithRecipients.Recipients...)
				if err != nil {
					return "", err
				}
				return "", nil
			},

			DashboardHelpers.TOPIC_CLOSE_CHILD: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				systemgeConnection := systemgeServer.GetConnection(message.GetPayload())
				if systemgeConnection == nil {
					return "", errors.New("Connection not found")
				}
				err := systemgeConnection.Close()
				if err != nil {
					return "", err
				}
				return helpers.IntToString(status.Stopped), nil
			},
			DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_SEQUENTIALLY_CHILD: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				systemgeConnection := systemgeServer.GetConnection(message.GetPayload())
				if systemgeConnection == nil {
					return "", errors.New("Connection not found")
				}
				err := systemgeConnection.StartMessageHandlingLoop(messageHandler, true)
				if err != nil {
					return "", err
				}
				return "", nil
			},
			DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_CONCURRENTLY_CHILD: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				systemgeConnection := systemgeServer.GetConnection(message.GetPayload())
				if systemgeConnection == nil {
					return "", errors.New("Connection not found")
				}
				err := systemgeConnection.StartMessageHandlingLoop(messageHandler, false)
				if err != nil {
					return "", err
				}
				return "", nil
			},
			DashboardHelpers.TOPIC_STOP_MESSAGE_HANDLING_LOOP_CHILD: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				systemgeConnection := systemgeServer.GetConnection(message.GetPayload())
				if systemgeConnection == nil {
					return "", errors.New("Connection not found")
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
					return "", errors.New("Connection not found")
				}
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
				return handleNextMessageResult.Marshal(), nil
			},
		},
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
		eventHandler,
	)
}
