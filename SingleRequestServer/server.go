package SingleRequestServer

import (
	"sync/atomic"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Dashboard"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
)

type SingleRequestServer struct {
	config          *Config.SingleRequestServer
	commandHandlers Commands.Handlers
	messageHandler  SystemgeConnection.MessageHandler
	systemgeServer  *SystemgeServer.SystemgeServer
	dashboardClient *Dashboard.Client

	// metrics
	invalidMessages atomic.Uint64

	succeededCommands atomic.Uint64
	failedCommands    atomic.Uint64

	succeededAsyncMessages atomic.Uint64
	failedAsyncMessages    atomic.Uint64

	succeededSyncMessages atomic.Uint64
	failedSyncMessages    atomic.Uint64
}

func NewSingleRequestServer(name string, config *Config.SingleRequestServer, commands Commands.Handlers, messageHandler SystemgeConnection.MessageHandler) *SingleRequestServer {
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
	server.systemgeServer = SystemgeServer.New(name, config.SystemgeServerConfig, server.onConnect, nil)
	if config.DashboardClientConfig != nil {
		server.dashboardClient = Dashboard.NewClient(name+"_dashboardClient", config.DashboardClientConfig, server.Start, server.Stop, server.RetrieveMetrics, server.GetStatus, commands)
		err := server.dashboardClient.Start()
		if err != nil {
			panic(Error.New("Failed to start dashboard client", err))
		}
	}
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
			server.failedCommands.Add(1)
			return Error.New("No commands available on this server", nil)
		}
		command := unmarshalCommandStruct(message.GetPayload())
		if command == nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Invalid command")
			server.failedCommands.Add(1)
			return Error.New("Invalid command", nil)
		}
		handler := server.commandHandlers[command.Command]
		if handler == nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Command not found")
			server.failedCommands.Add(1)
			return Error.New("Command not found", nil)
		}
		result, err := handler(command.Args)
		if err != nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, err.Error())
			server.failedCommands.Add(1)
			return Error.New("Command failed", err)
		}
		server.succeededCommands.Add(1)
		connection.SyncRequestBlocking(Message.TOPIC_SUCCESS, result)
		connection.Close()
		return nil
	case "async":
		if server.messageHandler == nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "No message handler available")
			server.failedAsyncMessages.Add(1)
			return Error.New("No message handler available on this server", nil)
		}
		asyncMessage, err := Message.Deserialize([]byte(message.GetPayload()), message.GetOrigin())
		if err != nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Failed to deserialize message")
			server.failedAsyncMessages.Add(1)
			return err
		}
		err = server.messageHandler.HandleAsyncMessage(connection, asyncMessage)
		if err != nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, err.Error())
			server.failedAsyncMessages.Add(1)
			return Error.New("Message handler failed", err)
		}
		server.succeededAsyncMessages.Add(1)
		connection.SyncRequestBlocking(Message.TOPIC_SUCCESS, "")
		connection.Close()
		return nil
	case "sync":
		if server.messageHandler == nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "No message handler available")
			server.failedSyncMessages.Add(1)
			return Error.New("No message handler available on this server", nil)
		}
		syncMessage, err := Message.Deserialize([]byte(message.GetPayload()), message.GetOrigin())
		if err != nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Failed to deserialize message")
			server.failedSyncMessages.Add(1)
			return err
		}
		payload, err := server.messageHandler.HandleSyncRequest(connection, syncMessage)
		if err != nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, err.Error())
			server.failedSyncMessages.Add(1)
			return Error.New("Message handler failed", err)
		}
		server.succeededSyncMessages.Add(1)
		connection.SyncRequestBlocking(Message.TOPIC_SUCCESS, payload)
		connection.Close()
		return nil
	default:
		server.invalidMessages.Add(1)
		connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Invalid topic")
		return Error.New("Invalid topic", nil)
	}
}

func (server *SingleRequestServer) Start() error {
	return server.systemgeServer.Start()
}

func (server *SingleRequestServer) Stop() error {
	return server.systemgeServer.Stop()
}

func (server *SingleRequestServer) GetStatus() int {
	return server.systemgeServer.GetStatus()
}

func (server *SingleRequestServer) GetMetrics() map[string]uint64 {
	return server.systemgeServer.GetMetrics()
}
func (server *SingleRequestServer) RetrieveMetrics() map[string]uint64 {
	return server.systemgeServer.RetrieveMetrics()
}
