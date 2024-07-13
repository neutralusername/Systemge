package Oauth2

import (
	"Systemge/Utilities"
	"time"
)

func (server *Server) GetSession(sessionId string) *Session {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.sessions[sessionId]
}

func (server *Server) addSession(session *Session) string {
	sessionId := ""
	server.mutex.Lock()
	for {
		sessionId = server.config.Randomizer.GenerateRandomString(32, Utilities.ALPHA_NUMERIC)
		if _, ok := server.sessions[sessionId]; !ok {
			server.sessions[sessionId] = session
			session.watchdog = time.AfterFunc(time.Duration(server.config.SessionLifetimeMs)*time.Millisecond, func() {
				server.removeSession(sessionId)
			})
			break
		}
	}
	server.mutex.Unlock()
	return sessionId
}

func (server *Server) removeSession(sessionId string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	delete(server.sessions, sessionId)
}
