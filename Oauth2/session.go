package Oauth2

type Session struct {
	keyValuePairs map[string]interface{}
}

func NewSession(keyValuePairs map[string]interface{}) *Session {
	return &Session{keyValuePairs: keyValuePairs}
}

func (session *Session) Get(key string) (interface{}, bool) {
	value, ok := session.keyValuePairs[key]
	return value, ok
}
