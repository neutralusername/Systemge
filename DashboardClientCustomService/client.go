package DashboardClientCustomService

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/DashboardClient"
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func New_(name string, config *Config.DashboardClient, startFunc func() error, stopFunc func() error, getStatusFunc func() int, getMetricsFunc func() map[string]map[string]uint64, commands Commands.Handlers) *DashboardClient.Client {
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
	)
}

func New(name string, config *Config.DashboardClient, customService customService, commands Commands.Handlers) *DashboardClient.Client {
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
		panic("config.EndpointConfig is nil")
	}
	if customService == nil {
		panic("customService is nil")
	}
	return DashboardClient.New(
		name,
		config,
		SystemgeConnection.NewTopicExclusiveMessageHandler(
			nil,
			SystemgeConnection.SyncMessageHandlers{
				DashboardHelpers.TOPIC_GET_STATUS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					return Helpers.IntToString(customService.GetStatus()), nil
				},
				DashboardHelpers.TOPIC_GET_METRICS: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					return Helpers.JsonMarshal(DashboardHelpers.ConvertToDashboardMetrics(customService.GetMetrics())), nil
				},
				DashboardHelpers.TOPIC_START: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					err := customService.Start()
					if err != nil {
						return "", err
					}
					return Helpers.IntToString(customService.GetStatus()), nil
				},
				DashboardHelpers.TOPIC_STOP: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					err := customService.Stop()
					if err != nil {
						return "", err
					}
					return Helpers.IntToString(customService.GetStatus()), nil
				},
				DashboardHelpers.TOPIC_COMMAND: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
					command, err := DashboardHelpers.UnmarshalCommand(message.GetPayload())
					if err != nil {
						return "", err
					}
					return commands.Execute(command.Command, command.Args)
				},
			},
			nil, nil,
			1000,
		),
		func() (string, error) {
			pageMarshalled, err := DashboardHelpers.NewPage(
				DashboardHelpers.NewCustomServiceClient(
					name,
					commands.GetKeyBoolMap(),
					customService.GetStatus(),
					DashboardHelpers.ConvertToDashboardMetrics(customService.GetMetrics()),
				),
				DashboardHelpers.CLIENT_TYPE_CUSTOMSERVICE,
			).Marshal()
			if err != nil {
				panic(err)
			}
			return string(pageMarshalled), nil
		},
	)
}
