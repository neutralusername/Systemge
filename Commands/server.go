package Commands

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
)

type CommandServer struct {
	config               *Config.CommandServer
	commandHandlers      Handlers
	SystemgeServerConfig *SystemgeServer.SystemgeServer `json:"systemgeServerConfig"` // *required*
}

func NewCommandServer(name string, config *Config.CommandServer, commands Handlers) *CommandServer {
	if config == nil {
		panic("Config is required")
	}
	if config.SystemgeServerConfig == nil {
		panic("SystemgeServerConfig is required")
	}
	if config.SystemgeServerConfig.ConnectionConfig == nil {
		panic("ConnectionConfig is required")
	}
	if config.SystemgeServerConfig.ListenerConfig == nil {
		panic("TcpServerConfig is required")
	}

	commandServer := &CommandServer{
		config:          config,
		commandHandlers: commands,
	}
	commandServer.SystemgeServerConfig = SystemgeServer.New(name, config.SystemgeServerConfig, commandServer.onConnect, nil)
	return commandServer
}

func (commandServer *CommandServer) onConnect(connection SystemgeConnection.SystemgeConnection) error {
	message, err := connection.GetNextMessage()
	if err != nil {
		return err
	}
	if message.GetTopic() != "command" {
		connection.SyncResponse(message, false, "Invalid topic")
		return Error.New("Invalid topic", nil)
	}
	command := UnmarshalCommandStruct(message.GetPayload())
	handler := commandServer.commandHandlers[command.Command]
	if handler == nil {
		connection.SyncResponse(message, false, "Command not found")
		return Error.New("Command not found", nil)
	}
	result, err := handler(command.Args)
	if err != nil {
		connection.SyncResponse(message, false, err.Error())
		return Error.New("Command failed", err)
	}
	connection.SyncResponse(message, true, result)
	return nil
}

func (commandServer *CommandServer) Start() error {
	return commandServer.SystemgeServerConfig.Start()
}

func (commandServer *CommandServer) Stop() error {
	return commandServer.SystemgeServerConfig.Stop()
}

func (commandServer *CommandServer) GetStatus() int {
	return commandServer.SystemgeServerConfig.GetStatus()
}

func (commandServer *CommandServer) GetMetrics() map[string]uint64 {
	return commandServer.SystemgeServerConfig.GetMetrics()
}
