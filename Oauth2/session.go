package Oauth2

type Session struct {
	keyValuePairs map[string]interface{}
}

func (session *Session) Get(key string) (interface{}, bool) {
	value, ok := session.keyValuePairs[key]
	return value, ok
}
