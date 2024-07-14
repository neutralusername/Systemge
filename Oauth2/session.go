package Oauth2

import (
	"time"
)

type session struct {
	keyValuePairs map[string]interface{}
	watchdog      *time.Timer
	identity      string
	sessionId     string
	expired       bool
}

func (session *session) Removed() bool {
	return session.watchdog == nil
}

func newSession(sessionId, identity string, keyValuePairs map[string]interface{}) *session {
	return &session{
		keyValuePairs: keyValuePairs,
		identity:      identity,
		sessionId:     sessionId,
	}
}

func (session *session) Get(key string) (interface{}, bool) {
	value, ok := session.keyValuePairs[key]
	return value, ok
}

func (server *Server) Refresh(session *session) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if session.Removed() {
		return
	}
	session.watchdog.Reset(time.Duration(server.config.SessionLifetimeMs) * time.Millisecond)
	session.expired = false
}

func (server *Server) Expire(session *session) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if session.Removed() {
		return
	}
	session.watchdog.Reset(0)
}

func (server *Server) stop(session *session) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if session.Removed() {
		return
	}
	session.watchdog.Stop()
}
