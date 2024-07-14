package Oauth2

import (
	"time"
)

type session struct {
	keyValuePairs map[string]interface{}
	watchdog      *time.Timer
}

func (session *session) Expired() bool {
	return session.watchdog == nil
}

func newSession(keyValuePairs map[string]interface{}) *session {
	return &session{keyValuePairs: keyValuePairs}
}

func (session *session) Get(key string) (interface{}, bool) {
	value, ok := session.keyValuePairs[key]
	return value, ok
}

func (server *Server) Refresh(session *session) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if session.Expired() {
		return
	}
	session.watchdog.Reset(time.Duration(server.config.SessionLifetimeMs) * time.Millisecond)
}

func (server *Server) Expire(session *session) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if session.Expired() {
		return
	}
	session.watchdog.Reset(0)
}

func (server *Server) stop(session *session) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if session.Expired() {
		return
	}
	session.watchdog.Stop()
}
