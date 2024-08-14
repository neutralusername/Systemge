package Oauth2Server

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Tools"
)

type Server struct {
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
		ServerConfig: config.ServerConfig,
	}, server.GetHTTPMessageHandlers())
	return server
}

func (server *Server) Start() {
	go func() {
		err := server.httpServer.Start()
		if err != nil {
			panic(err)
		}
	}()
	server.mutex.Lock()
	defer server.mutex.Unlock()
	server.sessionRequestChannel = make(chan *oauth2SessionRequest)
	go handleSessionRequests(server)
}

func (server *Server) Stop() {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	close(server.sessionRequestChannel)
	server.removeAllSessions()
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
