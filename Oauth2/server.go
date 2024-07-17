package Oauth2

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Http"
	"Systemge/Utilities"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
)

type Server struct {
	httpServer            *http.Server
	sessionRequestChannel chan *oauth2SessionRequest
	config                *Config.Oauth2

	sessions   map[string]*session
	identities map[string]*session
	mutex      sync.Mutex

	stopChannel chan string
	isStarted   bool
}

func New(config Config.Oauth2) (*Server, error) {
	if config.Randomizer == nil {
		config.Randomizer = Utilities.NewRandomizer(Utilities.GetSystemTime())
	}
	if config.TokenHandler == nil {
		return nil, Error.New("TokenHandler is required", nil)
	}
	if config.OAuth2Config == nil {
		return nil, Error.New("OAuth2Config is required", nil)
	}
	server := &Server{
		sessionRequestChannel: make(chan *oauth2SessionRequest),
		config:                &config,
		sessions:              make(map[string]*session),
		identities:            make(map[string]*session),
	}
	return server, nil
}

func (server *Server) Start() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Error.New("oauth2 server \""+server.config.Name+"\" is already started", nil)
	}
	server.httpServer = Http.New(server.config.Port, map[string]Http.RequestHandler{
		server.config.AuthPath:         server.oauth2Auth(),
		server.config.AuthCallbackPath: server.oauth2Callback(),
	})
	err := Http.Start(server.httpServer, "", "")
	if err != nil {
		return Error.New("failed to start oauth2 server \""+server.config.Name+"\"", err)
	}
	server.stopChannel = make(chan string)
	go handleSessionRequests(server)
	server.isStarted = true
	server.config.Logger.Info(Error.New("started oauth2 server \""+server.config.Name+"\"", nil).Error())
	return nil
}

func (server *Server) Stop() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if !server.isStarted {
		return Error.New("oauth2 server \""+server.config.Name+"\" is not started", nil)
	}
	err := Http.Stop(server.httpServer)
	if err != nil {
		return Error.New("failed to stop oauth2 server \""+server.config.Name+"\"", err)
	}
	server.httpServer = nil
	server.isStarted = false
	close(server.stopChannel)
	server.removeAllSessions()
	server.config.Logger.Info(Error.New("stopped oauth2 server \""+server.config.Name+"\"", nil).Error())
	return nil
}

func (server *Server) GetName() string {
	return server.config.Name
}

func (server *Server) GetLogger() *Utilities.Logger {
	return server.config.Logger
}

func (server *Server) GetOauth2Config() *oauth2.Config {
	return server.config.OAuth2Config
}

func handleSessionRequests(server *Server) {
	for {
		select {
		case sessionRequest := <-server.sessionRequestChannel:
			server.config.Logger.Info(Error.New("handling session request with access token \""+sessionRequest.token.AccessToken+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
			handleSessionRequest(server, sessionRequest)
		case <-server.stopChannel:
			server.config.Logger.Info(Error.New("stopped handling session requests on oauth2 server \""+server.config.Name+"\"", nil).Error())
			return
		}
	}
}

func handleSessionRequest(server *Server, sessionRequest *oauth2SessionRequest) {
	identity, keyValuePairs, err := server.config.TokenHandler(server.config.OAuth2Config, sessionRequest.token)
	if err != nil {
		sessionRequest.sessionChannel <- nil
		server.config.Logger.Warning(Error.New("failed handling session request for access token \""+sessionRequest.token.AccessToken+"\" on oauth2 server \""+server.config.Name+"\"", err).Error())
		return
	}
	if identity == "" {
		sessionRequest.sessionChannel <- nil
		server.config.Logger.Warning(Error.New("no session identity for access token \""+sessionRequest.token.AccessToken+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
		return
	}
	sessionRequest.sessionChannel <- server.getSessionForIdentity(identity, keyValuePairs)
}
