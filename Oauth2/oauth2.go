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

func (server *Server) Start() {
	go handleSessionRequests(server)
	Http.Start(server.httpServer, "", "")
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
		sessionId = server.randomizer.GenerateRandomString(32, Utilities.ALPHA_NUMERIC)
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
