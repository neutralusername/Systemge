package Oauth2Server

type session struct {
	identity  string
	sessionId string

	closeChannel chan bool
}

func newSession(sessionId, identity string) *session {
	return &session{
		identity:  identity,
		sessionId: sessionId,
	}
}

func (session *session) GetSessionId() string {
	return session.sessionId
}

func (session *session) GetIdentity() string {
	return session.identity
}
