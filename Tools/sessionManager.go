package Tools

import (
	"errors"
	"math"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
)

// can be theoretically used "recursively" to represent groups as identities and members as sessions. think about how to generalize the naming

type SessionManager struct {
	config *Config.SessionManager

	identities map[string]*Identity
	sessions   map[string]*Session

	onExpire         func(*Session)
	onCreate         func(*Session) error
	maxTotalSessions float64

	acceptSessions bool

	eventHandler Event.Handler

	mutex sync.RWMutex
}

func NewSessionManager(config *Config.SessionManager, onCreate func(*Session) error, onExpire func(*Session), eventHandler Event.Handler) *SessionManager {
	if config == nil {
		config = &Config.SessionManager{}
	}
	if config.SessionIdLength == 0 {
		config.SessionIdLength = 32
	}
	if config.SessionIdAlphabet == "" {
		config.SessionIdAlphabet = ALPHA_NUMERIC
	}
	return &SessionManager{
		config: config,

		sessions:   make(map[string]*Session),
		identities: make(map[string]*Identity),

		onCreate: onCreate,
		onExpire: onExpire,

		acceptSessions: true,

		maxTotalSessions: math.Pow(float64(len(config.SessionIdAlphabet)), float64(config.SessionIdLength)),

		eventHandler: eventHandler,
	}
}

func (manager *SessionManager) CreateSession(identityString string) (*Session, error) {
	manager.mutex.Lock()
	if !manager.acceptSessions {
		manager.mutex.Unlock()
		return nil, errors.New("session manager not accepting sessions")
	}
	if len(manager.sessions) >= int(manager.maxTotalSessions) {
		manager.mutex.Unlock()
		return nil, errors.New("max total sessions exceeded")
	}
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
	}
	sessionId := GenerateRandomString(manager.config.SessionIdLength, manager.config.SessionIdAlphabet)
	for {
		if _, ok := manager.sessions[sessionId]; !ok {
			break
		}
		sessionId = GenerateRandomString(manager.config.SessionIdLength, manager.config.SessionIdAlphabet)
	}
	session := &Session{
		id:            sessionId,
		identity:      identity,
		keyValuePairs: make(map[string]any),
	}
	identity.sessions[session.GetId()] = nil
	manager.sessions[session.GetId()] = nil
	manager.mutex.Unlock()

	if err := manager.onCreate(session); err != nil {
		manager.mutex.Lock()
		delete(identity.sessions, session.id)
		delete(manager.sessions, session.id)
		manager.mutex.Unlock()
		return nil, err
	} else {
		manager.mutex.Lock()
		manager.identities[identity.GetId()] = identity
		identity.sessions[session.GetId()] = session
		manager.sessions[session.GetId()] = session
		manager.mutex.Unlock()
	}

	session.timeout = NewTimeout(
		manager.config.SessionLifetimeMs,
		func() {
			manager.mutex.Lock()
			identity := session.identity
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
	if session, ok := manager.sessions[sessionId]; ok && session != nil {
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
		sessions := []*Session{}
		for _, session := range identity.sessions {
			if session != nil {
				sessions = append(sessions, session)
			}
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

// default is true. returns the new value
func (manager *SessionManager) ToggleAcceptSeccions() bool {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	manager.acceptSessions = !manager.acceptSessions
	return manager.acceptSessions
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
