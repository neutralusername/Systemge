package RPCServer

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
)

type RPCServer struct {
	config          *Config.CommandServer
	commandHandlers Commands.Handlers
	SystemgeServer  *SystemgeServer.SystemgeServer
}

func NewRPCServer(name string, config *Config.CommandServer, commands Commands.Handlers) *RPCServer {
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

	commandServer := &RPCServer{
		config:          config,
		commandHandlers: commands,
	}
	commandServer.SystemgeServer = SystemgeServer.New(name, config.SystemgeServerConfig, commandServer.onConnect, nil)
	return commandServer
}

func (commandServer *RPCServer) onConnect(connection SystemgeConnection.SystemgeConnection) error {
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

func (commandServer *RPCServer) Start() error {
	return commandServer.SystemgeServer.Start()
}

func (commandServer *RPCServer) Stop() error {
	return commandServer.SystemgeServer.Stop()
}

func (commandServer *RPCServer) GetStatus() int {
	return commandServer.SystemgeServer.GetStatus()
}

func (commandServer *RPCServer) GetMetrics() map[string]uint64 {
	return commandServer.SystemgeServer.GetMetrics()
}
func (commandServer *RPCServer) RetrieveMetrics() map[string]uint64 {
	return commandServer.SystemgeServer.RetrieveMetrics()
}
