package SingleRequestServer

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
)

type SingleRequestServer struct {
	config          *Config.RPCServer
	commandHandlers Commands.Handlers
	messageHandler  SystemgeConnection.MessageHandler
	SystemgeServer  *SystemgeServer.SystemgeServer
}

func NewSingleRequestServer(name string, config *Config.RPCServer, commands Commands.Handlers, messageHandler SystemgeConnection.MessageHandler) *SingleRequestServer {
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

	server := &SingleRequestServer{
		config:          config,
		commandHandlers: commands,
		messageHandler:  messageHandler,
	}
	server.SystemgeServer = SystemgeServer.New(name, config.SystemgeServerConfig, server.onConnect, nil)
	return server
}

func (server *SingleRequestServer) onConnect(connection SystemgeConnection.SystemgeConnection) error {
	message, err := connection.GetNextMessage()
	if err != nil {
		connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Failed to get message")
		return err
	}
	switch message.GetTopic() {
	case "command":
		if server.commandHandlers == nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "No commands available")
			return Error.New("No commands available on this server", nil)
		}
		command := UnmarshalCommandStruct(message.GetPayload())
		if command == nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Invalid command")
			return Error.New("Invalid command", nil)
		}
		handler := server.commandHandlers[command.Command]
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
	case "async":
		if server.messageHandler == nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "No message handler available")
			return Error.New("No message handler available on this server", nil)
		}
		asyncMessage, err := Message.Deserialize([]byte(message.GetPayload()), message.GetOrigin())
		if err != nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Failed to deserialize message")
			return err
		}
		err = server.messageHandler.HandleAsyncMessage(connection, asyncMessage)
		if err != nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, err.Error())
			return Error.New("Message handler failed", err)
		}
		connection.SyncRequestBlocking(Message.TOPIC_SUCCESS, "")
		connection.Close()
		return nil
	case "sync":
		if server.messageHandler == nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "No message handler available")
			return Error.New("No message handler available on this server", nil)
		}
		syncMessage, err := Message.Deserialize([]byte(message.GetPayload()), message.GetOrigin())
		if err != nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Failed to deserialize message")
			return err
		}
		payload, err := server.messageHandler.HandleSyncRequest(connection, syncMessage)
		if err != nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, err.Error())
			return Error.New("Message handler failed", err)
		}
		connection.SyncRequestBlocking(Message.TOPIC_SUCCESS, payload)
		connection.Close()
		return nil
	default:
		connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Invalid topic")
		return Error.New("Invalid topic", nil)
	}
}

func (server *SingleRequestServer) Start() error {
	return server.SystemgeServer.Start()
}

func (server *SingleRequestServer) Stop() error {
	return server.SystemgeServer.Stop()
}

func (server *SingleRequestServer) GetStatus() int {
	return server.SystemgeServer.GetStatus()
}

func (server *SingleRequestServer) GetMetrics() map[string]uint64 {
	return server.SystemgeServer.GetMetrics()
}
func (server *SingleRequestServer) RetrieveMetrics() map[string]uint64 {
	return server.SystemgeServer.RetrieveMetrics()
}
