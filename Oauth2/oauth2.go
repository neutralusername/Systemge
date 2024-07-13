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
	sessionRequestChannel chan *Http.Oauth2SessionRequest
	randomizer            *Utilities.Randomizer
	tokenHandler          func(*Server, *oauth2.Token) (map[string]interface{}, error)

	sessions map[string]*Identity
	mutex    sync.Mutex
}

type Identity struct {
	keyValuePairs map[string]interface{}
}

func NewServer(port int, authPath, authCallbackPath string, oAuthConfig *oauth2.Config, logger *Utilities.Logger, sessionRequestHandler func(*Server, *oauth2.Token) (map[string]interface{}, error)) *Server {
	server := &Server{
		logger:                logger,
		sessionRequestChannel: make(chan *Http.Oauth2SessionRequest),
		oauth2Config:          oAuthConfig,
		tokenHandler:          sessionRequestHandler,
		randomizer:            Utilities.NewRandomizer(Utilities.GetSystemTime()),

		sessions: make(map[string]*Identity),
	}
	server.oauth2State = server.randomizer.GenerateRandomString(16, Utilities.ALPHA_NUMERIC)
	server.httpServer = Http.New(port, map[string]Http.RequestHandler{
		authPath:         Http.Oauth2(server.oauth2Config, server.oauth2State),
		authCallbackPath: Http.Oauth2Callback(server.oauth2Config, server.oauth2State, server.logger, server.sessionRequestChannel, "http://127.0.0.1:8080/", "http://google.at"),
	})
	return server
}

func (server *Server) Start() {
	go handleSessionRequests(server)
	Http.Start(server.httpServer, "", "")
}

func handleSessionRequests(server *Server) {
	func() {
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
				server.sessions[sessionId] = &Identity{
					keyValuePairs: keyValuePairs,
				}
				break
			}
		}
		server.mutex.Unlock()
		sessionRequest.SessionIdChannel <- sessionId
	}()
}
