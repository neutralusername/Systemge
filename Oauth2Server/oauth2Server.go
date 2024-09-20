package Oauth2Server

import (
	"net/http"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type Server struct {
	name        string
	status      int
	statusMutex sync.Mutex

	sessionRequestChannel chan *oauth2SessionRequest
	config                *Config.Oauth2

	sessions   map[string]*session
	identities map[string]*session

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger

	randomizer *Tools.Randomizer
	httpServer *HTTPServer.HTTPServer

	mutex sync.Mutex
}

func New(name string, config *Config.Oauth2, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList) *Server {
	if config.TokenHandler == nil {
		panic("TokenHandler is required")
	}
	if config.OAuth2Config == nil {
		panic("OAuth2Config is required")
	}
	server := &Server{
		name:          name,
		config:        config,
		sessions:      make(map[string]*session),
		identities:    make(map[string]*session),
		infoLogger:    Tools.NewLogger("[Info: \"Oauth2\"]", config.InfoLoggerPath),
		warningLogger: Tools.NewLogger("[Warning: \"Oauth2\"]", config.WarningLoggerPath),
		httpServer:    nil,

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

func (server *Server) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STOPPED {
		return Event.New("Server is not in stopped state", nil)
	}
	server.status = Status.PENDING
	server.sessionRequestChannel = make(chan *oauth2SessionRequest)
	err := server.httpServer.Start()
	if err != nil {
		server.status = Status.STOPPED
		close(server.sessionRequestChannel)
		server.sessionRequestChannel = nil
		return err
	}
	go handleSessionRequests(server)
	server.status = Status.STARTED
	return nil
}

func (server *Server) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STARTED {
		return Event.New("Server is not in started state", nil)
	}
	server.status = Status.PENDING
	server.httpServer.Stop()
	close(server.sessionRequestChannel)

	server.mutex.Lock()
	for _, session := range server.sessions {
		session.watchdog.Stop()
		session.watchdog = nil
		delete(server.sessions, session.sessionId)
		delete(server.identities, session.identity)
	}
	server.mutex.Unlock()

	server.status = Status.STOPPED
	return nil
}

func handleSessionRequests(server *Server) {
	for sessionRequest := range server.sessionRequestChannel {
		if sessionRequest == nil {
			return
		}
		if infoLogger := server.infoLogger; infoLogger != nil {
			infoLogger.Log(Event.New("Handling session request with access token \""+sessionRequest.token.AccessToken+"\"", nil).Error())
		}
		handleSessionRequest(server, sessionRequest)
	}
}

func handleSessionRequest(server *Server, sessionRequest *oauth2SessionRequest) {
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
