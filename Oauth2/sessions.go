package Oauth2

import (
	"Systemge/Utilities"
	"time"
)

func (server *Server) GetSession(sessionId string) *session {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.sessions[sessionId]
}

func (server *Server) addSession(session *session) string {
	sessionId := ""
	server.mutex.Lock()
	for {
		sessionId = server.config.Randomizer.GenerateRandomString(32, Utilities.ALPHA_NUMERIC)
		if _, ok := server.sessions[sessionId]; !ok {
			server.sessions[sessionId] = session
			session.watchdog = time.AfterFunc(time.Duration(server.config.SessionLifetimeMs)*time.Millisecond, func() {
				server.mutex.Lock()
				defer server.mutex.Unlock()
				if session.Expired() {
					return
				}
				session.watchdog = nil
				delete(server.sessions, sessionId)
				server.config.Logger.Info("removed session \"" + sessionId + "\"")
			})
			break
		}
	}
	server.config.Logger.Info("created session \"" + sessionId + "\"")
	server.mutex.Unlock()
	return sessionId
}
