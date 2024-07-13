package Oauth2

import "time"

type Session struct {
	keyValuePairs map[string]interface{}
	watchdog      *time.Timer
}

func newSession(keyValuePairs map[string]interface{}) *Session {
	return &Session{keyValuePairs: keyValuePairs}
}

func (session *Session) Get(key string) (interface{}, bool) {
	value, ok := session.keyValuePairs[key]
	return value, ok
}
