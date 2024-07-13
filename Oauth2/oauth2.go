package Oauth2

import (
	"Systemge/Http"
	"Systemge/Utilities"
	"net/http"

	"golang.org/x/oauth2"
)

type Server struct {
	httpServer   *http.Server
	sessions     map[string]*Identity
	oauth2Config *oauth2.Config
	oauth2State  string
	logger       *Utilities.Logger
	tokenChannel chan *oauth2.Token
	randomizer   *Utilities.Randomizer
	tokenHandler func(*oauth2.Token)
}

type Identity struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

func NewServer(port int, authPath, authCallbackPath string, oAuthConfig *oauth2.Config, logger *Utilities.Logger, tokenHandler func(*oauth2.Token)) *Server {
	server := &Server{
		sessions:     make(map[string]*Identity),
		logger:       logger,
		tokenChannel: make(chan *oauth2.Token),
		oauth2Config: oAuthConfig,
		randomizer:   Utilities.NewRandomizer(Utilities.GetSystemTime()),
	}
	server.oauth2State = server.randomizer.GenerateRandomString(16, Utilities.ALPHA_NUMERIC)
	server.httpServer = Http.New(port, map[string]Http.RequestHandler{
		authPath:         Http.Oauth2(server.oauth2Config, server.oauth2State),
		authCallbackPath: Http.Oauth2Callback(server.oauth2Config, server.oauth2State, server.logger, server.tokenChannel, "localhost:8080/", "http://google.at"),
	})
	return server
}

func (server *Server) Start() {
	go func() {
		token := <-server.tokenChannel
		server.tokenHandler(token)
	}()
	Http.Start(server.httpServer, "", "")
}
