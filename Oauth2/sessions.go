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

func (server *Server) createSession(identity string, keyValuePairs map[string]interface{}) (session *session) {
	for {
		sessionId := server.config.Randomizer.GenerateRandomString(32, Utilities.ALPHA_NUMERIC)
		if _, ok := server.sessions[sessionId]; !ok {
			session = newSession(sessionId, identity, keyValuePairs)
			server.sessions[sessionId] = session
			server.identities[identity] = session
			session.watchdog = time.AfterFunc(time.Duration(server.config.SessionLifetimeMs)*time.Millisecond, server.removeSession(session))
			break
		}
	}
	return
}

func (server *Server) removeSession(session *session) func() {
	return func() {
		session.expired = true
		server.mutex.Lock()
		defer server.mutex.Unlock()
		if !session.expired {
			return
		}
		if session.watchdog == nil {
			return
		}
		session.watchdog = nil
		delete(server.sessions, session.sessionId)
		delete(server.identities, session.identity)
		server.config.Logger.Info(Error.New("removed session \""+session.sessionId+"\" with identity \""+session.identity+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
	}
}

func (server *Server) removeSessions() {
	for _, session := range server.sessions {
		session.watchdog.Stop()
		session.watchdog = nil
		delete(server.sessions, session.sessionId)
		delete(server.identities, session.identity)
	}
}
