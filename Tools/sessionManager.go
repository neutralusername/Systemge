package Tools

import (
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
)

type SessionManager struct {
	config *Config.SessionManager

	identities map[string]*Identity
	sessions   map[string]*Session

	onExpire func(*Session)
	onCreate func(*Session) error

	eventHandler Event.Handler

	mutex sync.RWMutex
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
		if manager.config.MaxIdentities > 0 && uint32(len(manager.identities)) >= manager.config.MaxIdentities {
			manager.mutex.Unlock()
			return nil, errors.New("max identities exceeded")
		}
		identity = &Identity{
			id:       identityString,
			sessions: make(map[string]*Session),
		}
		manager.identities[identity.id] = identity
	}
	sessionId := GenerateRandomString(manager.config.SessionIdLength, ALPHA_NUMERIC)
	for _, ok := manager.sessions[sessionId]; ok; {
		sessionId = GenerateRandomString(manager.config.SessionIdLength, ALPHA_NUMERIC)
	}
	session := &Session{
		id:       sessionId,
		identity: identity,
	}
	identity.sessions[session.GetId()] = session
	manager.sessions[session.GetId()] = session
	manager.mutex.Unlock()

	if err := manager.onCreate(session); err != nil {
		manager.mutex.Lock()
		delete(identity.sessions, session.id)
		delete(manager.sessions, session.id)
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

func (session *Session) Set(key string, value any) {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	session.keyValuePairs[key] = value
}

func (session *Session) Get(key string) (any, bool) {
	session.mutex.RLock()
	defer session.mutex.RUnlock()
	val, ok := session.keyValuePairs[key]
	return val, ok
}

func (session *Session) Remove(key string) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	if _, ok := session.keyValuePairs[key]; !ok {
		return errors.New("key not found")
	}
	delete(session.keyValuePairs, key)
	return nil
}

func (session *Session) GetMap() map[string]any {
	return session.keyValuePairs
}

func (session *Session) GetId() string {
	return session.id
}

func (session *Session) GetIdentity() string {
	return session.identity.GetId()
}

// is nil until the onCreate is finished
func (session *Session) GetTimeout() *Timeout {
	return session.timeout
}

func NewSessionManager(config *Config.SessionManager, onCreate func(*Session) error, onExpire func(*Session), eventHandler Event.Handler) *SessionManager {
	if config == nil {
		config = &Config.SessionManager{}
	}
	if config.SessionIdLength == 0 {
		config.SessionIdLength = 32
	}
	return &SessionManager{
		config: config,

		sessions:   make(map[string]*Session),
		identities: make(map[string]*Identity),

		onCreate: onCreate,
		onExpire: onExpire,

		eventHandler: eventHandler,
	}
}

type Identity struct {
	id       string
	sessions map[string]*Session
}

func (identity *Identity) GetId() string {
	return identity.id
}

type Session struct {
	id            string
	identity      *Identity
	mutex         sync.RWMutex
	keyValuePairs map[string]any

	timeout *Timeout
}
