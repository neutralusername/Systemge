package SingleRequestServer

import (
	"encoding/json"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Dashboard"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
	"github.com/neutralusername/Systemge/Tools"
)

type Server struct {
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

func NewSingleRequestServer(name string, config *Config.SingleRequestServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, commands Commands.Handlers, messageHandler SystemgeConnection.MessageHandler) *Server {
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

	server := &Server{
		config:          config,
		commandHandlers: commands,
		messageHandler:  messageHandler,
	}
	server.systemgeServer = SystemgeServer.New(name, config.SystemgeServerConfig, whitelist, blacklist, server.onConnect, nil)
	if config.DashboardClientConfig != nil {
		defaultCommands := server.GetDefaultCommands()
		for key, value := range commands {
			defaultCommands[key] = value
		}
		server.dashboardClient = Dashboard.NewClient(name+"_dashboardClient", config.DashboardClientConfig, server.Start, server.Stop, server.RetrieveMetrics, server.GetStatus, defaultCommands)
		err := server.StartDashboard()
		if err != nil {
			panic(Error.New("Failed to start dashboard client", err))
		}
	}
	return server
}

func (server *Server) onConnect(connection SystemgeConnection.SystemgeConnection) error {
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
		connection.AsyncMessage(Message.TOPIC_SUCCESS, "")
		err = server.messageHandler.HandleAsyncMessage(connection, asyncMessage)
		if err != nil {
			server.failedAsyncMessages.Add(1)
			return Error.New("Message handler failed", err)
		}
		server.succeededAsyncMessages.Add(1)
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

func (server *Server) Start() error {
	return server.systemgeServer.Start()
}

func (server *Server) Stop() error {
	return server.systemgeServer.Stop()
}

func (server *Server) StartDashboard() error {
	if server.dashboardClient == nil {
		return Error.New("No dashboard client available", nil)
	}
	return server.dashboardClient.Start()
}

func (server *Server) StopDashboard() error {
	if server.dashboardClient == nil {
		return Error.New("No dashboard client available", nil)
	}
	return server.dashboardClient.Stop()
}

func (server *Server) GetStatus() int {
	return server.systemgeServer.GetStatus()
}

func (server *Server) GetMetrics() map[string]uint64 {
	metrics := map[string]uint64{}
	metrics["invalid_messages"] = server.GetInvalidMessages()
	metrics["succeeded_commands"] = server.GetSucceededCommands()
	metrics["failed_commands"] = server.GetFailedCommands()
	metrics["succeeded_async_messages"] = server.GetSucceededAsyncMessages()
	metrics["failed_async_messages"] = server.GetFailedAsyncMessages()
	metrics["succeeded_sync_messages"] = server.GetSucceededSyncMessages()
	metrics["failed_sync_messages"] = server.GetFailedSyncMessages()
	serverMetrics := server.systemgeServer.GetMetrics()
	for key, value := range serverMetrics {
		metrics[key] = value
	}
	return metrics
}
func (server *Server) RetrieveMetrics() map[string]uint64 {
	metrics := map[string]uint64{}
	metrics["invalid_messages"] = server.RetrieveInvalidMessages()
	metrics["succeeded_commands"] = server.RetrieveSucceededCommands()
	metrics["failed_commands"] = server.RetrieveFailedCommands()
	metrics["succeeded_async_messages"] = server.RetrieveSucceededAsyncMessages()
	metrics["failed_async_messages"] = server.RetrieveFailedAsyncMessages()
	metrics["succeeded_sync_messages"] = server.RetrieveSucceededSyncMessages()
	metrics["failed_sync_messages"] = server.RetrieveFailedSyncMessages()
	serverMetrics := server.systemgeServer.RetrieveMetrics()
	for key, value := range serverMetrics {
		metrics[key] = value
	}
	return metrics
}

func (server *Server) GetInvalidMessages() uint64 {
	return server.invalidMessages.Load()
}
func (server *Server) RetrieveInvalidMessages() uint64 {
	return server.invalidMessages.Swap(0)
}

func (server *Server) GetSucceededCommands() uint64 {
	return server.succeededCommands.Load()
}
func (server *Server) RetrieveSucceededCommands() uint64 {
	return server.succeededCommands.Swap(0)
}

func (server *Server) GetFailedCommands() uint64 {
	return server.failedCommands.Load()
}
func (server *Server) RetrieveFailedCommands() uint64 {
	return server.failedCommands.Swap(0)
}

func (server *Server) GetSucceededAsyncMessages() uint64 {
	return server.succeededAsyncMessages.Load()
}
func (server *Server) RetrieveSucceededAsyncMessages() uint64 {
	return server.succeededAsyncMessages.Swap(0)
}

func (server *Server) GetFailedAsyncMessages() uint64 {
	return server.failedAsyncMessages.Load()
}
func (server *Server) RetrieveFailedAsyncMessages() uint64 {
	return server.failedAsyncMessages.Swap(0)
}

func (server *Server) GetSucceededSyncMessages() uint64 {
	return server.succeededSyncMessages.Load()
}
func (server *Server) RetrieveSucceededSyncMessages() uint64 {
	return server.succeededSyncMessages.Swap(0)
}

func (server *Server) GetFailedSyncMessages() uint64 {
	return server.failedSyncMessages.Load()
}
func (server *Server) RetrieveFailedSyncMessages() uint64 {
	return server.failedSyncMessages.Swap(0)
}

func (server *Server) GetDefaultCommands() Commands.Handlers {
	commands := Commands.Handlers{}
	commands["start"] = func(args []string) (string, error) { //overriding the start command from the systemgeServer
		err := server.Start()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["stop"] = func(args []string) (string, error) { //overriding the stop command from the systemgeServer
		err := server.Stop()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["getStatus"] = func(args []string) (string, error) { //overriding the getStatus command from the systemgeServer
		return Status.ToString(server.GetStatus()), nil
	}
	commands["getMetrics"] = func(args []string) (string, error) { //overriding the getMetrics command from the systemgeServer
		metrics := server.GetMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Error.New("Failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	commands["retrieveMetrics"] = func(args []string) (string, error) { //overriding the retrieveMetrics command from the systemgeServer
		metrics := server.RetrieveMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Error.New("Failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	serverCommands := server.systemgeServer.GetDefaultCommands()
	for key, value := range serverCommands {
		commands["systemgeServer_"+key] = value
	}
	return commands
}
