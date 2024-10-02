package WebsocketServer

import (
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
	"github.com/neutralusername/Systemge/WebsocketListener"
)

type WebsocketServer struct {
	name string

	status      int
	statusMutex sync.Mutex

	stopChannel chan bool
	waitGroup   sync.WaitGroup

	instanceId string
	sessionId  string

	config *Config.WebsocketServer

	websocketListener *WebsocketListener.WebsocketListener

	sessionManager *Tools.SessionManager
	topicManager   *Tools.TopicManager

	eventHandler Event.Handler
}

func New(name string, config *Config.WebsocketServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, eventHandler Event.Handler) (*WebsocketServer, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.SessionManagerConfig == nil {
		return nil, errors.New("config.SessionManagerConfig is nil")
	}
	if config.WebsocketConnectionConfig == nil {
		return nil, errors.New("config.WebsocketConnectionConfig is nil")
	}
	if config.WebsocketListenerConfig == nil {
		return nil, errors.New("config.WebsocketListenerConfig is nil")
	}
	websocketListener, err := WebsocketListener.New(name+"_websocketListener", config.WebsocketListenerConfig, whitelist, blacklist, nil)
	if err != nil {
		return nil, err
	}

	server := &WebsocketServer{
		name:       name,
		instanceId: Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

		sessionManager: Tools.NewSessionManager(name+"_sessionManager", config.SessionManagerConfig, nil, nil),

		config: config,

		websocketListener: websocketListener,
	}
	/* 	topicHandlers := make(Tools.TopicHandlers)
	   	for topic, handler := range messageHandlers {
	   		topicHandlers[topic] = server.toTopicHandler(handler)
	   	}
	   	server.topicManager = Tools.NewTopicManager(config.TopicManagerConfig, topicHandlers, nil) */

	return server, nil
}

func (server *WebsocketServer) GetName() string {
	return server.name
}

func (server *WebsocketServer) GetStatus() int {
	return server.status
}

func (server *WebsocketServer) GetInstanceId() string {
	return server.instanceId
}

func (server *WebsocketServer) GetSessionId() string {
	return server.sessionId
}

func (server *WebsocketServer) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(server.GetServerContext())
	server.eventHandler(event)
	return event
}
func (server *WebsocketServer) GetServerContext() Event.Context {
	return Event.Context{
		Event.ServiceType:   Event.WebsocketServer,
		Event.ServiceName:   server.name,
		Event.ServiceStatus: Status.ToString(server.status),
		Event.Function:      Event.GetCallerFuncName(2),
		Event.InstanceId:    server.instanceId,
		Event.SessionId:     server.sessionId,
	}
}

func (server *WebsocketServer) toTopicHandler(handler WebsocketClient.WebsocketMessageHandler) Tools.TopicHandler {
	return func(args ...any) (any, error) {
		websocketConnection := args[0].(*WebsocketClient.WebsocketClient)
		message := args[1].(*Message.Message)

		if server.eventHandler != nil {
			if event := server.onEvent(Event.NewInfo(
				Event.HandlingMessage,
				"handling websocketConnection message",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance: Event.HandleMessage,
					Event.Identity:     websocketConnection.GetName(),
					Event.Address:      websocketConnection.GetAddress(),
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
			)); !event.IsInfo() {
				return nil, errors.New("event cancelled")
			}
		}

		err := handler(websocketConnection, message)
		if err != nil {
			if server.eventHandler != nil {
				server.onEvent(Event.NewWarningNoOption(
					Event.HandleMessageFailed,
					err.Error(),
					Event.Context{
						Event.Circumstance: Event.HandleMessage,
						Event.Identity:     websocketConnection.GetName(),
						Event.Address:      websocketConnection.GetAddress(),
						Event.Topic:        message.GetTopic(),
						Event.Payload:      message.GetPayload(),
						Event.SyncToken:    message.GetSyncToken(),
					},
				))
			}
			return nil, err
		}

		if server.eventHandler != nil {
			server.onEvent(Event.NewInfoNoOption(
				Event.HandledMessage,
				"handled websocketConnection message",
				Event.Context{
					Event.Circumstance: Event.HandleMessage,
					Event.Identity:     websocketConnection.GetName(),
					Event.Address:      websocketConnection.GetAddress(),
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
			))
		}

		return nil, nil
	}
}
