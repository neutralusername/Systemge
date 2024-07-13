package Oauth2

import (
	"Systemge/Http"
	"Systemge/Utilities"
	"net/http"

	"golang.org/x/oauth2"
)

type Server struct {
	httpServer            *http.Server
	sessions              map[string]*Identity
	oauth2Config          *oauth2.Config
	oauth2State           string
	logger                *Utilities.Logger
	sessionRequestChannel chan *Http.Oauth2SessionRequest
	randomizer            *Utilities.Randomizer
	sessionRequestHandler func(*Server, *Http.Oauth2SessionRequest)
}

type Identity struct {
	AccessToken  string
	RefreshToken string
}

func NewServer(port int, authPath, authCallbackPath string, oAuthConfig *oauth2.Config, logger *Utilities.Logger, tokenHandler func(*Server, *Http.Oauth2SessionRequest)) *Server {
	server := &Server{
		sessions:              make(map[string]*Identity),
		logger:                logger,
		sessionRequestChannel: make(chan *Http.Oauth2SessionRequest),
		oauth2Config:          oAuthConfig,
		sessionRequestHandler: tokenHandler,
		randomizer:            Utilities.NewRandomizer(Utilities.GetSystemTime()),
	}
	server.oauth2State = server.randomizer.GenerateRandomString(16, Utilities.ALPHA_NUMERIC)
	server.httpServer = Http.New(port, map[string]Http.RequestHandler{
		authPath:         Http.Oauth2(server.oauth2Config, server.oauth2State),
		authCallbackPath: Http.Oauth2Callback(server.oauth2Config, server.oauth2State, server.logger, server.sessionRequestChannel, "http://127.0.0.1:8080/", "http://google.at"),
	})
	return server
}

func (server *Server) Start() {
	go func() {
		sessionRequest := <-server.sessionRequestChannel
		server.sessionRequestHandler(server, sessionRequest)
	}()
	Http.Start(server.httpServer, "", "")
}
