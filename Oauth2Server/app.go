package Oauth2Server

import (
	"net/http"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type Server struct {
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

func New(config *Config.Oauth2) *Server {
	if config.TokenHandler == nil {
		panic("TokenHandler is required")
	}
	if config.OAuth2Config == nil {
		panic("OAuth2Config is required")
	}
	server := &Server{
		config:        config,
		sessions:      make(map[string]*session),
		identities:    make(map[string]*session),
		infoLogger:    Tools.NewLogger("[Info: \"Oauth2\"]", config.InfoLoggerPath),
		warningLogger: Tools.NewLogger("[Warning: \"Oauth2\"]", config.WarningLoggerPath),
		httpServer:    nil,

		randomizer: Tools.NewRandomizer(config.RandomizerSeed),
	}
	server.httpServer = HTTPServer.New(&Config.HTTPServer{
		TcpListenerConfig: config.TcpListenerConfig,
	}, map[string]http.HandlerFunc{
		server.config.AuthPath:         server.oauth2Auth(),
		server.config.AuthCallbackPath: server.oauth2AuthCallback(),
	})
	return server
}

func (server *Server) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STOPPED {
		return Error.New("Server is not in stopped state", nil)
	}
	server.status = Status.PENDING
	err := server.httpServer.Start()
	if err != nil {
		server.status = Status.STOPPED
		return err
	}
	server.mutex.Lock()
	defer server.mutex.Unlock()
	server.sessionRequestChannel = make(chan *oauth2SessionRequest)
	go handleSessionRequests(server)
	server.status = Status.STARTED
	return nil
}

func (server *Server) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STARTED {
		return Error.New("Server is not in started state", nil)
	}
	server.status = Status.PENDING
	server.mutex.Lock()
	defer server.mutex.Unlock()
	server.httpServer.Stop()
	close(server.sessionRequestChannel)
	server.removeAllSessions()
	server.status = Status.STOPPED
	return nil
}

func handleSessionRequests(server *Server) {
	for sessionRequest := range server.sessionRequestChannel {
		if sessionRequest == nil {
			return
		}
		if infoLogger := server.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Handling session request with access token \""+sessionRequest.token.AccessToken+"\"", nil).Error())
		}
		handleSessionRequest(server, sessionRequest)
	}
}

func handleSessionRequest(server *Server, sessionRequest *oauth2SessionRequest) {
	identity, keyValuePairs, err := server.config.TokenHandler(server.config.OAuth2Config, sessionRequest.token)
	if err != nil {
		sessionRequest.sessionChannel <- nil
		if warningLogger := server.warningLogger; warningLogger != nil {
			warningLogger.Log(Error.New("Failed handling session request for access token \""+sessionRequest.token.AccessToken+"\"", err).Error())
		}
		return
	}
	if identity == "" {
		sessionRequest.sessionChannel <- nil
		if warningLogger := server.warningLogger; warningLogger != nil {
			warningLogger.Log(Error.New("No session identity for access token \""+sessionRequest.token.AccessToken+"\"", nil).Error())
		}
		return
	}
	sessionRequest.sessionChannel <- server.getSessionForIdentity(identity, keyValuePairs)
}
