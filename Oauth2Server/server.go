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
)

type Server struct {
	name        string
	status      int
	statusMutex sync.Mutex

	instanceId string
	sessionId  string

	sessionRequestChannel chan *oauth2SessionRequest
	config                *Config.Oauth2

	sessions   map[string]*session
	identities map[string]*session

	randomizer *Tools.Randomizer
	httpServer *HTTPServer.HTTPServer

	eventHandler Event.Handler

	mutex sync.Mutex
}

func New(name string, config *Config.Oauth2, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, eventHandler Event.Handler) *Server {
	if config.TokenHandler == nil {
		panic("TokenHandler is required")
	}
	if config.OAuth2Config == nil {
		panic("OAuth2Config is required")
	}
	server := &Server{
		name:       name,
		config:     config,
		sessions:   make(map[string]*session),
		identities: make(map[string]*session),
		httpServer: nil,

		instanceId: Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

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
