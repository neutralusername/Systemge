package Tools

import (
	"errors"
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
	onCreate func(*Session) error

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
	identity *Identity

	timeout *Timeout
}

func (session *Session) GetId() string {
	return session.id
}

func (session *Session) GetIdentity() string {
	return session.identity.GetId()
}

func (session *Session) GetTimeout() *Timeout {
	return session.timeout
}

func NewSessionManager(config *Config.SessionManager, onCreate func(*Session) error, onExpire func(*Session), eventHandler Event.Handler) *SessionManager {
	return &SessionManager{
		config: config,

		sessions:   make(map[string]*Session),
		identities: make(map[string]*Identity),

		onCreate: onCreate,
		onExpire: onExpire,

		eventHandler: eventHandler,
	}
}

func (manager *SessionManager) CreateSession(identityString string) (*Session, error) {
	manager.mutex.Lock()

	identity, ok := manager.identities[identityString]
	if ok {
		if manager.config.MaxSessionsPerIdentity > 0 && uint32(len(identity.sessions)) >= manager.config.MaxSessionsPerIdentity {
			manager.mutex.Unlock()
			return nil, errors.New("max sessions per identity exceeded")
		}
	} else {
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

	if err := manager.onCreate(session); err != nil {
		if len(identity.sessions) == 0 {
			delete(manager.identities, identity.GetId())
		}
		manager.mutex.Unlock()
		return nil, err
	}

	session.timeout = NewTimeout(
		manager.config.SessionLifetimeMs,
		func() {
			manager.mutex.Lock()
			delete(identity.sessions, session.id)
			delete(manager.sessions, session.id)
			if len(identity.sessions) == 0 {
				delete(manager.identities, identity.GetId())
			}
			manager.mutex.Unlock()

			manager.onExpire(session)
		},
		false,
	)
	identity.sessions[session.GetId()] = session
	manager.sessions[session.GetId()] = session
	manager.mutex.Unlock()

	return session, nil
}

func (manager *SessionManager) GetSession(sessionId string) *Session {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	if session, ok := manager.sessions[sessionId]; ok {
		return session
	}
	return nil
}

func (manager *SessionManager) GetIdentities() []string {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	identities := make([]string, 0, len(manager.identities))
	for identity := range manager.identities {
		identities = append(identities, identity)
	}
	return identities
}

func (manager *SessionManager) GetSessions(identityString string) []*Session {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	if identity, ok := manager.identities[identityString]; ok {
		sessions := make([]*Session, 0, len(identity.sessions))
		for _, session := range identity.sessions {
			sessions = append(sessions, session)
		}
		return sessions
	}
	return nil
}

func (manager *SessionManager) HasActiveSession(identityString string) bool {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	if _, ok := manager.identities[identityString]; ok {
		return true
	}
	return false
}
