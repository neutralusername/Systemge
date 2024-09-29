package SingleRequestServer

import (
	"errors"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
	"github.com/neutralusername/Systemge/Tools"
)

type Server struct {
	config          *Config.SingleRequestServer
	commandHandlers Commands.Handlers
	messageHandler  SystemgeConnection.MessageHandler
	systemgeServer  *SystemgeServer.SystemgeServer

	// metrics
	invalidRequests atomic.Uint64

	succeededCommands atomic.Uint64
	failedCommands    atomic.Uint64

	succeededAsyncMessages atomic.Uint64
	failedAsyncMessages    atomic.Uint64

	succeededSyncMessages atomic.Uint64
	failedSyncMessages    atomic.Uint64
}

func NewSingleRequestServer(name string, config *Config.SingleRequestServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, commands Commands.Handlers, messageHandler SystemgeConnection.MessageHandler, eventHandler Event.Handler) (*Server, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}
	if config.SystemgeServerConfig == nil {
		return nil, errors.New("systemgeServerConfig is required")
	}
	if config.SystemgeServerConfig.TcpSystemgeConnectionConfig == nil {
		return nil, errors.New("tcpSystemgeConnectionConfig is required")
	}
	if config.SystemgeServerConfig.TcpSystemgeListenerConfig == nil {
		return nil, errors.New("tcpSystemgeListenerConfig is required")
	}

	server := &Server{
		config:          config,
		commandHandlers: commands,
		messageHandler:  messageHandler,
	}
	systemgeServer, err := SystemgeServer.New(name, config.SystemgeServerConfig, whitelist, blacklist, func(event *Event.Event) {
		eventHandler(event)
		switch event.GetEvent() {
		case Event.HandledAcception:
			event.SetWarning()
			clientName, ok := event.GetContextValue(Event.ClientName)
			if !ok {
				break
			}
			systemgeConnection := server.systemgeServer.GetConnection(clientName)
			err := server.onConnect(systemgeConnection)
			if err != nil {
				break
			}
		}
	})
	if err != nil {
		return nil, err
	}
	server.systemgeServer = systemgeServer
	return server, nil
}

func (server *Server) onConnect(connection SystemgeConnection.SystemgeConnection) error {
	message, err := connection.RetrieveNextMessage()
	if err != nil {
		connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Failed to get message")
		return err
	}
	switch message.GetTopic() {
	case "command":
		if server.commandHandlers == nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "No commands available")
			server.failedCommands.Add(1)
			return errors.New("no commands available on this server")
		}
		command := unmarshalCommandStruct(message.GetPayload())
		if command == nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Invalid command")
			server.failedCommands.Add(1)
			return errors.New("invalid command")
		}
		handler := server.commandHandlers[command.Command]
		if handler == nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Command not found")
			server.failedCommands.Add(1)
			return errors.New("Command not found")
		}
		result, err := handler(command.Args)
		if err != nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, err.Error())
			server.failedCommands.Add(1)
			return err
		}
		server.succeededCommands.Add(1)
		connection.SyncRequestBlocking(Message.TOPIC_SUCCESS, result)
		connection.Close()
		return nil
	case "async":
		if server.messageHandler == nil {
			server.failedAsyncMessages.Add(1)
			return errors.New("no message handler available on this server")
		}
		asyncMessage, err := Message.Deserialize([]byte(message.GetPayload()), message.GetOrigin())
		if err != nil {
			server.failedAsyncMessages.Add(1)
			return err
		}
		err = server.messageHandler.HandleAsyncMessage(connection, asyncMessage)
		if err != nil {
			server.failedAsyncMessages.Add(1)
			return err
		}
		server.succeededAsyncMessages.Add(1)
		connection.Close()
		return nil
	case "sync":
		if server.messageHandler == nil {
			connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "No message handler available")
			server.failedSyncMessages.Add(1)
			return errors.New("no message handler available on this server")
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
			return err
		}
		server.succeededSyncMessages.Add(1)
		connection.SyncRequestBlocking(Message.TOPIC_SUCCESS, payload)
		connection.Close()
		return nil
	default:
		server.invalidRequests.Add(1)
		connection.SyncRequestBlocking(Message.TOPIC_FAILURE, "Invalid topic")
		return errors.New("invalid topic")
	}
}

func (server *Server) Start() error {
	return server.systemgeServer.Start()
}

func (server *Server) Stop() error {
	return server.systemgeServer.Stop()
}

func (server *Server) GetStatus() int {
	return server.systemgeServer.GetStatus()
}
