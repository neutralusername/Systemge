package Tools

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
)

type SessionManager struct {
	config *Config.SessionManager

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

type Session struct {
	id       string
	lifetime uint64
	identity *Identity
}

func (session *Session) GetId() string {
	return session.id
}

func (session *Session) GetIdentity() string {
	return session.identity.GetId()
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
		config: config,

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
		lifetime: manager.config.SessionLifetimeMs,
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

func (manager *SessionManager) GetSession(sessionId string) *Session {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	if session, ok := manager.sessions[sessionId]; ok {
		return session
	}
	return nil
}

func (manager *SessionManager) GetSessions(identityString string) []*Session {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	identity, ok := manager.identities[identityString]
	if !ok {
		return nil
	}
	sessions := make([]*Session, 0, len(identity.sessions))
	for _, session := range identity.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}
