package Commands

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
)

type CommandServer struct {
	config          *Config.CommandServer
	commandHandlers Handlers
	SystemgeServer  *SystemgeServer.SystemgeServer
}

// essentially a remote procedure call (RPC) server
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
	commandServer.SystemgeServer = SystemgeServer.New(name, config.SystemgeServerConfig, commandServer.onConnect, nil)
	return commandServer
}

func (commandServer *CommandServer) onConnect(connection SystemgeConnection.SystemgeConnection) error {
	message, err := connection.GetNextMessage()
	if err != nil {
		connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Failed to get message")
		return err
	}
	if message.GetTopic() != "command" {
		connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Invalid topic")
		return Error.New("Invalid topic", nil)
	}
	command := UnmarshalCommandStruct(message.GetPayload())
	if command == nil {
		connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Invalid command")
		return Error.New("Invalid command", nil)
	}
	handler := commandServer.commandHandlers[command.Command]
	if handler == nil {
		connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Command not found")
		return Error.New("Command not found", nil)
	}
	result, err := handler(command.Args)
	if err != nil {
		connection.SyncRequestBlocking(Message.TOPIC_FAILURE, err.Error())
		return Error.New("Command failed", err)
	}
	connection.SyncRequestBlocking(Message.TOPIC_SUCCESS, result)
	connection.Close()
	return nil
}

func (commandServer *CommandServer) Start() error {
	return commandServer.SystemgeServer.Start()
}

func (commandServer *CommandServer) Stop() error {
	return commandServer.SystemgeServer.Stop()
}

func (commandServer *CommandServer) GetStatus() int {
	return commandServer.SystemgeServer.GetStatus()
}

func (commandServer *CommandServer) GetMetrics() map[string]uint64 {
	return commandServer.SystemgeServer.GetMetrics()
}
func (commandServer *CommandServer) RetrieveMetrics() map[string]uint64 {
	return commandServer.SystemgeServer.RetrieveMetrics()
}
