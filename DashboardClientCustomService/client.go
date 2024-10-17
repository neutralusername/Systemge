package DashboardClientCustomService

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

func New_(name string, config *configs.DashboardClient, startFunc func() error, stopFunc func() error, getStatusFunc func() int, getMetricsFunc func() Metrics.MetricsTypes, commands Commands.Handlers, eventHandler Event.Handler) (*DashboardClient.Client, error) {
	if startFunc == nil {
		panic("startFunc is nil")
	}
	if stopFunc == nil {
		panic("stopFunc is nil")
	}
	if getStatusFunc == nil {
		panic("getStatusFunc is nil")
	}
	if getMetricsFunc == nil {
		panic("getMetricsFunc is nil")
	}
	return New(
		name,
		config,
		&customServiceStruct{
			startFunc:      startFunc,
			stopFunc:       stopFunc,
			getStatusFunc:  getStatusFunc,
			getMetricsFunc: getMetricsFunc,
		},
		commands,
		eventHandler,
	)
}

func New(name string, config *configs.DashboardClient, customService customService, commands Commands.Handlers, eventHandler Event.Handler) (*DashboardClient.Client, error) {
	if config == nil {
		panic("config is nil")
	}
	if name == "" {
		panic("config.Name is empty")
	}
	if config.TcpSystemgeConnectionConfig == nil {
		panic("config.ConnectionConfig is nil")
	}
	if config.TcpClientConfig == nil {
		panic("config.TcpClientConfig is nil")
	}
	if customService == nil {
		panic("customService is nil")
	}
	return DashboardClient.New(
		name,
		config,
		nil,
		SystemgeConnection.SyncMessageHandlers{
			DashboardHelpers.TOPIC_GET_STATUS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				return helpers.IntToString(customService.GetStatus()), nil
			},
			DashboardHelpers.TOPIC_GET_METRICS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				return helpers.JsonMarshal(customService.GetMetrics()), nil
			},

			DashboardHelpers.TOPIC_COMMAND: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				command, err := DashboardHelpers.UnmarshalCommand(message.GetPayload())
				if err != nil {
					return "", err
				}
				return commands.Execute(command.Command, command.Args)
			},
			DashboardHelpers.TOPIC_START: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				err := customService.Start()
				if err != nil {
					return "", err
				}
				return helpers.IntToString(customService.GetStatus()), nil
			},
			DashboardHelpers.TOPIC_STOP: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
				err := customService.Stop()
				if err != nil {
					return "", err
				}
				return helpers.IntToString(customService.GetStatus()), nil
			},
		},
		func() (string, error) {
			pageMarshalled, err := DashboardHelpers.NewPage(
				DashboardHelpers.NewCustomServiceClient(
					name,
					commands.GetKeyBoolMap(),
					customService.GetStatus(),
					DashboardHelpers.NewDashboardMetrics(customService.GetMetrics()),
				),
				DashboardHelpers.CLIENT_TYPE_CUSTOMSERVICE,
			).Marshal()
			if err != nil {
				panic(err)
			}
			return string(pageMarshalled), nil
		},
		eventHandler,
	)
}
