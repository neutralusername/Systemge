package DashboardClientCommands

import (
	"github.com/neutralusername/systemge/Commands"
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/DashboardClient"
	"github.com/neutralusername/systemge/DashboardHelpers"
	"github.com/neutralusername/systemge/Event"
	"github.com/neutralusername/systemge/Message"
	"github.com/neutralusername/systemge/Metrics"
	"github.com/neutralusername/systemge/SystemgeConnection"
	"github.com/neutralusername/systemge/helpers"
)

// frontend not implemented nor is this tested (use DashboardClientCustomService for now)
func New(name string, config *configs.DashboardClient, getMetricsFunc func() map[string]*Metrics.Metrics, commands Commands.Handlers, eventHandler Event.Handler) (*DashboardClient.Client, error) {
	var metrics DashboardHelpers.DashboardMetrics
	if getMetricsFunc != nil {
		metrics = DashboardHelpers.NewDashboardMetrics(getMetricsFunc())
	} else {
		metrics = DashboardHelpers.DashboardMetrics{}
	}
	return DashboardClient.New(
		name,
		config,
		nil,
		SystemgeConnection.SyncMessageHandlers{
			DashboardHelpers.TOPIC_GET_METRICS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				if getMetricsFunc == nil {
					return "", nil
				}
				return helpers.JsonMarshal(getMetricsFunc()), nil
			},

			DashboardHelpers.TOPIC_COMMAND: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				command, err := DashboardHelpers.UnmarshalCommand(message.GetPayload())
				if err != nil {
					return "", err
				}
				return commands.Execute(command.Command, command.Args)
			},
		},

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
		eventHandler,
	)
}
