package DashboardClientCommands

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/DashboardClient"
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func New(name string, config *Config.DashboardClient, getMetricsFunc func() map[string]map[string]uint64, commands Commands.Handlers) *DashboardClient.Client {
	var metrics map[string]map[string]*DashboardHelpers.MetricsEntry = make(map[string]map[string]*DashboardHelpers.MetricsEntry)
	if getMetricsFunc != nil {
		metrics = DashboardHelpers.ConvertToDashboardMetrics(getMetricsFunc())
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
				DashboardHelpers.TOPIC_GET_METRICS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					if getMetricsFunc == nil {
						return "", nil
					}
					return Helpers.JsonMarshal(DashboardHelpers.ConvertToDashboardMetrics(getMetricsFunc())), nil
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
