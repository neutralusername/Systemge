package Tools

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
)

type SessionManager struct {
	sessions   map[string]*Session
	identities map[string]*Identity

	onExpire func(*Session)
	onCreate func(*Session)

	eventHandler Event.Handler

	mutex sync.RWMutex
}

type Identity struct {
	id       string
	sessions map[string]*Session
}

func (identity *Identity) GetId() string {
	return identity.id
}

func (identity *Identity) GetSessions() []*Session {
	sessions := make([]*Session, len(identity.sessions))
	i := 0
	for _, session := range identity.sessions {
		sessions[i] = session
		i++
	}
	return sessions
}

type Session struct {
	id       string
	identity *Identity
}

func (session *Session) GetId() string {
	return session.id
}

func (session *Session) GetIdentity() *Identity {
	return session.identity
}

func (session *Session) ExpireSession() {

}

func (session *Session) RefreshSession() {

}

func (session *Session) SetLifetime(lifetimeMs uint64) {

}

func (session *Session) GetRemainingLifetime() uint64 {

}

func NewSessionManager(config *Config.SessionManager, onCreate func(*Session), onExpire func(*Session), eventHandler Event.Handler) *SessionManager {
	return &SessionManager{
		sessions:   make(map[string]*Session),
		identities: make(map[string]*Identity),

		onCreate: onCreate,
		onExpire: onExpire,

		eventHandler: eventHandler,
	}
}

func (manager *SessionManager) CreateSession(identityString string) *Session {
	manager.mutex.Lock()
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
	manager.mutex.Unlock()

	manager.onCreate(session)
	go manager.handleSessionLifetime(session)

	return session
}
func (manager *SessionManager) handleSessionLifetime(session *Session) {

}
