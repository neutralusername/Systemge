package SingleRequestServer

import (
	"errors"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	Server1 "github.com/neutralusername/Systemge/Server"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

type Server struct {
	name           string
	config         *Config.SingleRequestServer
	messageHandler SystemgeConnection.MessageHandler
	systemgeServer *Server1.Server

	eventHandler Event.Handler

	// metrics

	succeededAsyncMessages atomic.Uint64
	failedAsyncMessages    atomic.Uint64

	succeededSyncMessages atomic.Uint64
	failedSyncMessages    atomic.Uint64
}

func NewSingleRequestServer(name string, config *Config.SingleRequestServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, messageHandler SystemgeConnection.MessageHandler, eventHandler Event.Handler) (*Server, error) {
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
	if messageHandler == nil {
		return nil, errors.New("messageHandler is required")
	}

	server := &Server{
		name:           name,
		config:         config,
		messageHandler: messageHandler,
		eventHandler:   eventHandler,
	}
	systemgeServer, err := Server1.New(name, config.SystemgeServerConfig, whitelist, blacklist, func(event *Event.Event) {
		if eventHandler != nil {
			eventHandler(event)
		}

		switch event.GetEvent() {
		case Event.HandledAcception:
			event.SetWarning()
			clientName, ok := event.GetContextValue(Event.ClientName)
			if !ok {
				server.onEvent(Event.NewErrorNoOption(
					Event.ContextDoesNotExist,
					"client name context does not exist",
					Event.Context{
						Event.Circumstance: Event.SingleRequestServerRequest,
						Event.IdentityType: Event.SystemgeConnection,
					},
				))
				return
			}
			systemgeConnection := server.systemgeServer.GetConnection(clientName)
			if systemgeConnection == nil {
				server.onEvent(Event.NewErrorNoOption(
					Event.SessionDoesNotExist,
					"client does not exist",
					Event.Context{
						Event.Circumstance: Event.SingleRequestServerRequest,
						Event.IdentityType: Event.SystemgeConnection,
						Event.ClientName:   clientName,
					},
				))
				return
			}
			message, err := systemgeConnection.RetrieveNextMessage()
			if err != nil {
				server.onEvent(Event.NewWarningNoOption(
					Event.ReceivingFromChannelFailed,
					err.Error(),
					Event.Context{
						Event.Circumstance: Event.SingleRequestServerRequest,
						Event.IdentityType: Event.SystemgeConnection,
						Event.ClientName:   clientName,
					},
				))
				return
			}
			if message.GetSyncToken() == "" {
				err = server.messageHandler.HandleAsyncMessage(systemgeConnection, message)
				if err != nil {
					server.failedAsyncMessages.Add(1)
					server.onEvent(Event.NewWarningNoOption(
						Event.HandleMessageFailed,
						err.Error(),
						Event.Context{
							Event.Circumstance: Event.SingleRequestServerRequest,
							Event.HandlerType:  Event.AsyncMessage,
							Event.IdentityType: Event.SystemgeConnection,
							Event.ClientName:   clientName,
						},
					))
					return
				}
				server.succeededAsyncMessages.Add(1)
			} else {
				payload, err := server.messageHandler.HandleSyncRequest(systemgeConnection, message)
				if err != nil {
					systemgeConnection.SyncResponse(message, false, err.Error())
					server.failedSyncMessages.Add(1)
					server.onEvent(Event.NewWarningNoOption(
						Event.HandleMessageFailed,
						err.Error(),
						Event.Context{
							Event.Circumstance: Event.SingleRequestServerRequest,
							Event.HandlerType:  Event.AsyncMessage,
							Event.IdentityType: Event.SystemgeConnection,
							Event.ClientName:   clientName,
						},
					))
					return
				}
				server.succeededSyncMessages.Add(1)
				systemgeConnection.SyncResponse(message, true, payload)
			}
		}
	})
	if err != nil {
		return nil, err
	}
	server.systemgeServer = systemgeServer
	return server, nil
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

func (server *Server) onEvent(event *Event.Event) *Event.Event {
	if server.eventHandler != nil {
		server.eventHandler(event)
	}
	return event
}
func (server *Server) GetContext() Event.Context {
	return Event.Context{
		Event.ServiceType:       Event.SingleRequestServer,
		Event.ServiceName:       server.name,
		Event.ServiceStatus:     Status.ToString(server.systemgeServer.GetStatus()),
		Event.Function:          Event.GetCallerFuncName(2),
		Event.ServiceInstanceId: server.systemgeServer.GetInstanceId(),
		Event.SessionId:         server.systemgeServer.GetSessionId(),
	}
}
