package Oauth2Server

import (
	"net/http"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/HTTPServer"
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

func (server *Server) handleSessionRequests() {
	for sessionRequest := range server.sessionRequestChannel {
		if sessionRequest == nil {
			return
		}
		if infoLogger := server.infoLogger; infoLogger != nil {
			infoLogger.Log(Event.New("Handling session request with access token \""+sessionRequest.token.AccessToken+"\"", nil).Error())
		}
		server.handleSessionRequest(sessionRequest)
	}
}

func (server *Server) handleSessionRequest(sessionRequest *oauth2SessionRequest) {
	identity, keyValuePairs, err := server.config.TokenHandler(server.config.OAuth2Config, sessionRequest.token)
	if err != nil {
		sessionRequest.sessionChannel <- nil
		if warningLogger := server.warningLogger; warningLogger != nil {
			warningLogger.Log(Event.New("Failed handling session request for access token \""+sessionRequest.token.AccessToken+"\"", err).Error())
		}
		return
	}
	if identity == "" {
		sessionRequest.sessionChannel <- nil
		if warningLogger := server.warningLogger; warningLogger != nil {
			warningLogger.Log(Event.New("No session identity for access token \""+sessionRequest.token.AccessToken+"\"", nil).Error())
		}
		return
	}
	sessionRequest.sessionChannel <- server.getSessionForIdentity(identity, keyValuePairs)
}
