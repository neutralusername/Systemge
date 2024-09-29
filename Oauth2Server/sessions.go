package Oauth2Server

import (
	"time"

	"github.com/neutralusername/Systemge/Tools"
)

func (server *Server) GetSession(sessionId string) *session {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.sessions[sessionId]
}

func (server *Server) Expire(session *session) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if session.watchdog == nil {
		return
	}
	session.watchdog.Reset(0)
}

func (server *Server) getSessionForIdentity(identity string) *session {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	session := server.identities[identity]
	if session == nil {
		session = server.createSession(identity)
	} else {
		session.watchdog.Reset(time.Duration(server.config.SessionLifetimeMs) * time.Millisecond)
		session.expired = false
	}
	return session
}

func (server *Server) createSession(identity string) (session *session) {
	for {
		sessionId := server.randomizer.GenerateRandomString(32, Tools.ALPHA_NUMERIC)
		if _, ok := server.sessions[sessionId]; !ok {
			session = newSession(sessionId, identity)
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
	}
}
