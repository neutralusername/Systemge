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

func (server *Server) getSessionForIdentity(identity string, keyValuePairs map[string]interface{}) *session {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	session := server.identities[identity]
	if session == nil {
		session = server.createSession(identity, keyValuePairs)
		server.config.Logger.Info(Error.New("added session \""+session.sessionId+"\" with identity \""+session.identity+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
	} else {
		session.watchdog.Reset(time.Duration(server.config.SessionLifetimeMs) * time.Millisecond)
		session.expired = false
		server.config.Logger.Info(Error.New("refreshed session \""+session.sessionId+"\" with identity \""+session.identity+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
	}
	return session
}

func (server *Server) createSession(identity string, keyValuePairs map[string]interface{}) (session *session) {
	for {
		sessionId := server.config.Randomizer.GenerateRandomString(32, Utilities.ALPHA_NUMERIC)
		if _, ok := server.sessions[sessionId]; !ok {
			session = newSession(sessionId, identity, keyValuePairs)
			server.sessions[sessionId] = session
			server.identities[identity] = session
			session.watchdog = time.AfterFunc(time.Duration(server.config.SessionLifetimeMs)*time.Millisecond, server.getRemoveSessionFunc(session))
			break
		}
	}
	return
}

func (server *Server) getRemoveSessionFunc(session *session) func() {
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

func (server *Server) removeAllSessions() {
	for _, session := range server.sessions {
		session.watchdog.Stop()
		session.watchdog = nil
		delete(server.sessions, session.sessionId)
		delete(server.identities, session.identity)
	}
}
