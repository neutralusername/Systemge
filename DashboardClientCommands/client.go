package DashboardClientCommands

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/DashboardClient"
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Metrics"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
)

func New(name string, config *Config.DashboardClient, getMetricsFunc func() map[string]*Metrics.Metrics, commands Commands.Handlers) *DashboardClient.Client {
	var metrics DashboardHelpers.DashboardMetrics
	if getMetricsFunc != nil {
		metrics = DashboardHelpers.NewDashboardMetrics(getMetricsFunc())
	} else {
		metrics = DashboardHelpers.DashboardMetrics{}
	}
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
				DashboardHelpers.TOPIC_GET_METRICS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					if getMetricsFunc == nil {
						return "", nil
					}
					return Helpers.JsonMarshal(DashboardHelpers.NewDashboardMetrics(getMetricsFunc())), nil
				},
			},
			nil, nil,
			1000,
		),

		func() (string, error) {
			pageMarshalled, err := DashboardHelpers.NewPage(
				DashboardHelpers.NewCommandClient(
					name,
					commands.GetKeyBoolMap(),
					metrics,
				),
				DashboardHelpers.CLIENT_TYPE_COMMAND,
			).Marshal()
			if err != nil {
				panic(err)
			}
			return string(pageMarshalled), nil
		},
	)
}
