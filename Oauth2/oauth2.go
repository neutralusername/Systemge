package Oauth2

import (
	"Systemge/Error"
	"Systemge/Http"
	"Systemge/Utilities"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
)

type Server struct {
	httpServer            *http.Server
	oauth2Config          *oauth2.Config
	oauth2State           string
	logger                *Utilities.Logger
	sessionRequestChannel chan *Oauth2SessionRequest
	randomizer            *Utilities.Randomizer
	tokenHandler          func(*Server, *oauth2.Token) (map[string]interface{}, error)

	sessions map[string]*Session
	mutex    sync.Mutex
}

type ServerConfig struct {
	Port                  int
	AuthPath              string
	AuthCallbackPath      string
	OAuth2Config          *oauth2.Config
	Logger                *Utilities.Logger
	SessionRequestHandler func(*Server, *oauth2.Token) (map[string]interface{}, error)
}

func (oauth2ServerConfig *ServerConfig) New() *Server {
	server := &Server{
		logger:                oauth2ServerConfig.Logger,
		sessionRequestChannel: make(chan *Oauth2SessionRequest),
		oauth2Config:          oauth2ServerConfig.OAuth2Config,
		tokenHandler:          oauth2ServerConfig.SessionRequestHandler,
		randomizer:            Utilities.NewRandomizer(Utilities.GetSystemTime()),

		sessions: make(map[string]*Session),
	}
	server.oauth2State = server.randomizer.GenerateRandomString(16, Utilities.ALPHA_NUMERIC)
	server.httpServer = Http.New(oauth2ServerConfig.Port, map[string]Http.RequestHandler{
		oauth2ServerConfig.AuthPath:         Oauth2(server.oauth2Config, server.oauth2State),
		oauth2ServerConfig.AuthCallbackPath: Oauth2Callback(server.oauth2Config, server.oauth2State, server.logger, server.sessionRequestChannel, "http://127.0.0.1:8080/", "http://google.at"),
	})
	return server
}

func (server *Server) Start() {
	go handleSessionRequests(server)
	Http.Start(server.httpServer, "", "")
}

func (server *Server) GetSession(sessionId string) *Session {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.sessions[sessionId]
}

func handleSessionRequests(server *Server) {
	sessionRequest := <-server.sessionRequestChannel
	keyValuePairs, err := server.tokenHandler(server, sessionRequest.Token)
	if err != nil {
		sessionRequest.SessionIdChannel <- ""
		server.logger.Warning(Error.New("failed handling session request", err).Error())
		return
	}
	sessionId := ""
	server.mutex.Lock()
	for {
		sessionId = server.randomizer.GenerateRandomString(16, Utilities.ALPHA_NUMERIC)
		if _, ok := server.sessions[sessionId]; !ok {
			server.sessions[sessionId] = &Session{
				keyValuePairs: keyValuePairs,
			}
			break
		}
	}
	server.mutex.Unlock()
	sessionRequest.SessionIdChannel <- sessionId
}
