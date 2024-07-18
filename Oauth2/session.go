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

func newSession(sessionId, identity string, keyValuePairs map[string]interface{}) *session {
	return &session{
		keyValuePairs: keyValuePairs,
		identity:      identity,
		sessionId:     sessionId,
	}
}

func (session *session) GetSessionId() string {
	return session.sessionId
}

func (session *session) GetIdentity() string {
	return session.identity
}

func (session *session) Get(key string) (interface{}, bool) {
	value, ok := session.keyValuePairs[key]
	return value, ok
}

func (server *Server) Expire(session *session) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if session.watchdog == nil {
		return
	}
	session.watchdog.Reset(0)
}
