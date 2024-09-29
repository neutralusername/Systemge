package Oauth2Server

import (
	"net/http"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
	"golang.org/x/oauth2"
)

type TokenHandler func(*oauth2.Config, *oauth2.Token) error

type Server struct {
	name        string
	status      int
	statusMutex sync.Mutex

	instanceId string
	sessionId  string

	config *Config.Oauth2

	tokenHandler TokenHandler

	randomizer *Tools.Randomizer
	httpServer *HTTPServer.HTTPServer

	eventHandler Event.Handler
}

func New(name string, config *Config.Oauth2, tokenHandler TokenHandler, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, eventHandler Event.Handler) *Server {
	if tokenHandler == nil {
		panic("TokenHandler is required")
	}
	if config.OAuth2Config == nil {
		panic("OAuth2Config is required")
	}
	server := &Server{
		name:       name,
		config:     config,
		httpServer: nil,

		instanceId: Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

		tokenHandler: tokenHandler,

		eventHandler: eventHandler,

		randomizer: Tools.NewRandomizer(config.RandomizerSeed),
	}
	server.httpServer = HTTPServer.New(name+"_httpServer",
		&Config.HTTPServer{
			TcpServerConfig: config.TcpServerConfig,
		},
		whitelist, blacklist,
		map[string]http.HandlerFunc{
			server.config.AuthPath:         server.oauth2Auth(),
			server.config.AuthCallbackPath: server.oauth2AuthCallback(),
		},
		eventHandler,
	)
	return server
}

func (server *Server) GetName() string {
	return server.name
}

func (server *Server) GetStatus() int {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	return server.status
}

func (server *Server) GetInstanceId() string {
	return server.instanceId
}

func (server *Server) GetSessionId() string {
	return server.sessionId
}

func (server *Server) onEvent(event *Event.Event) *Event.Event {
	if server.eventHandler != nil {
		server.eventHandler(event)
	}
	return event
}
func (server *Server) GetContext() Event.Context {
	return Event.Context{
		Event.ServiceType:   Event.SingleRequestServer,
		Event.ServiceName:   server.name,
		Event.ServiceStatus: Status.ToString(server.GetStatus()),
		Event.Function:      Event.GetCallerFuncName(2),
		Event.InstanceId:    server.GetInstanceId(),
		Event.SessionId:     server.GetSessionId(),
	}
}
