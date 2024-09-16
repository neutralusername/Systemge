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
	var metrics map[string]*Metrics.Metrics
	if getMetricsFunc != nil {
		metrics = getMetricsFunc()
	} else {
		metrics = map[string]*Metrics.Metrics{}
	}
	Metrics.Merge(metrics, systemgeServer.GetMetrics())
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
					var metrics map[string]*Metrics.Metrics
					if getMetricsFunc != nil {
						metrics = getMetricsFunc()
					} else {
						metrics = map[string]*Metrics.Metrics{}
					}
					Metrics.Merge(metrics, systemgeServer.GetMetrics())
					return Helpers.JsonMarshal(DashboardHelpers.NewDashboardMetrics(metrics)), nil
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
