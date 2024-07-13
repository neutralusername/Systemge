package Oauth2

import (
	"Systemge/Error"
	"Systemge/Http"
	"Systemge/Utilities"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

type Server struct {
	httpServer            *http.Server
	sessionRequestChannel chan *Oauth2SessionRequest
	config                *Config

	sessions map[string]*Session
	mutex    sync.Mutex
}

func (server *Server) Start() {
	go handleSessionRequests(server)
	Http.Start(server.httpServer, "", "")
}

func (server *Server) GetOauth2Config() *oauth2.Config {
	return server.config.OAuth2Config
}

func handleSessionRequests(server *Server) {
	sessionRequest := <-server.sessionRequestChannel
	keyValuePairs, err := server.config.TokenHandler(server, sessionRequest.Token)
	if err != nil {
		sessionRequest.SessionIdChannel <- ""
		server.config.Logger.Warning(Error.New("failed handling session request", err).Error())
		return
	}
	sessionRequest.SessionIdChannel <- server.addSession(newSession(keyValuePairs))
}

func (server *Server) addSession(session *Session) string {
	sessionId := ""
	server.mutex.Lock()
	for {
		sessionId = server.config.Randomizer.GenerateRandomString(32, Utilities.ALPHA_NUMERIC)
		if _, ok := server.sessions[sessionId]; !ok {
			server.sessions[sessionId] = session
			session.watchdog = time.AfterFunc(time.Duration(server.config.SessionLifetimeMs)*time.Millisecond, func() {
				server.mutex.Lock()
				delete(server.sessions, sessionId)
				server.mutex.Unlock()
			})
			break
		}
	}
	server.mutex.Unlock()
	return sessionId
}
