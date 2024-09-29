package Tools

import (
	"github.com/neutralusername/Systemge/Constants"
)

type SessionManager struct {
	sessions   map[string]*Session
	identities map[string]*Identity

	onExpire func(*Session)
	onCreate func(*Session)
}

type Identity struct {
	id       string
	sessions map[string]*Session
}

type Session struct {
	id       string
	identity *Identity
}

func NewSessionManager(onExpire func(*Session), onCreate func(*Session)) *SessionManager {
	return &SessionManager{
		sessions:   make(map[string]*Session),
		identities: make(map[string]*Identity),

		onExpire: onExpire,
		onCreate: onCreate,
	}
}

func (manager *SessionManager) CreateSession(identityString string) *Session {
	identity, ok := manager.identities[identityString]
	if !ok {
		identity = &Identity{
			id:       identityString,
			sessions: make(map[string]*Session),
		}

		manager.identities[identity.id] = identity
	}

	sessionId := GenerateRandomString(Constants.SessionIdLength, ALPHA_NUMERIC)
	for _, ok := manager.sessions[sessionId]; ok; {
		sessionId = GenerateRandomString(Constants.SessionIdLength, ALPHA_NUMERIC)
	}
	session := &Session{
		id:       sessionId,
		identity: identity,
	}

	identity.sessions[session.id] = session
	manager.sessions[session.id] = session

	manager.onCreate(session)

	return session
}
