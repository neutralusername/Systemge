package DashboardClientCommands

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/DashboardClient"
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func New(name string, config *Config.DashboardClient, commands Commands.Handlers) *DashboardClient.Client {
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
	return DashboardClient.New(
		name,
		config,
		SystemgeConnection.NewTopicExclusiveMessageHandler(
			nil,
			SystemgeConnection.SyncMessageHandlers{
				Message.TOPIC_EXECUTE_COMMAND: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
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
			return string(DashboardHelpers.NewIntroduction(
				DashboardHelpers.NewCommandClient(
					name,
					commands.GetKeys(),
				),
			).Marshal()), nil
		},
	)
}
