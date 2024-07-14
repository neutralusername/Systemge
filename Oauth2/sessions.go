package Oauth2

import (
	"Systemge/Error"
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
	defer server.mutex.Unlock()
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
				server.config.Logger.Info(Error.New("removed session \""+sessionId+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
			})
			break
		}
	}
	server.config.Logger.Info(Error.New("added session \""+sessionId+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
	return sessionId
}
