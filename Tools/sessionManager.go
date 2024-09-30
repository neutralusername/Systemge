package Tools

import (
	"errors"
	"math"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
)

type SessionManager struct {
	config           *Config.SessionManager
	maxTotalSessions float64

	sessionMutex sync.RWMutex
	identities   map[string]*Identity
	sessions     map[string]*Session

	onExpire func(*Session)
	onCreate func(*Session) error

	acceptSessions      bool
	waitgroup           sync.WaitGroup
	acceptSessionsMutex sync.Mutex

	eventHandler Event.Handler
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
	manager.sessionMutex.Lock()
	if !manager.acceptSessions {
		manager.sessionMutex.Unlock()
		return nil, errors.New("session manager not accepting sessions")
	}

	manager.waitgroup.Add(1)
	defer manager.waitgroup.Done()

	if len(manager.sessions) >= int(manager.maxTotalSessions) {
		manager.sessionMutex.Unlock()
		return nil, errors.New("max total sessions exceeded")
	}
	identity, ok := manager.identities[identityString]
	if ok {
		if manager.config.MaxSessionsPerIdentity > 0 && uint32(len(identity.sessions)) >= manager.config.MaxSessionsPerIdentity {
			manager.sessionMutex.Unlock()
			return nil, errors.New("max sessions per identity exceeded")
		}
	} else {
		if manager.config.MaxIdentities > 0 && uint32(len(manager.identities)) >= manager.config.MaxIdentities {
			manager.sessionMutex.Unlock()
			return nil, errors.New("max identities exceeded")
		}
		identity = &Identity{
			id:       identityString,
			sessions: make(map[string]*Session),
		}
		manager.identities[identity.GetId()] = identity
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
	manager.sessionMutex.Unlock()

	if err := manager.onCreate(session); err != nil {
		manager.sessionMutex.Lock()
		delete(identity.sessions, session.id)
		delete(manager.sessions, session.id)
		manager.sessionMutex.Unlock()
		return nil, err
	} else {
		manager.sessionMutex.Lock()
		delete(identity.sessions, session.id)
		delete(manager.sessions, session.id)
		if len(identity.sessions) == 0 {
			delete(manager.identities, identity.GetId())
		}
		manager.sessionMutex.Unlock()
	}

	session.timeout = NewTimeout(
		manager.config.SessionLifetimeMs,
		func() {
			manager.sessionMutex.Lock()
			delete(identity.sessions, session.id)
			delete(manager.sessions, session.id)
			if len(identity.sessions) == 0 {
				delete(manager.identities, identity.GetId())
			}
			manager.sessionMutex.Unlock()

			manager.onExpire(session)
		},
		false,
	)

	return session, nil
}

func (manager *SessionManager) IsAcceptingSessions() bool {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()
	return manager.acceptSessions
}

func (manager *SessionManager) AcceptSessions() error {
	manager.acceptSessionsMutex.Lock()
	defer manager.acceptSessionsMutex.Unlock()

	manager.sessionMutex.Lock()
	if manager.acceptSessions {
		manager.sessionMutex.Unlock()
		return errors.New("session manager already accepting sessions")
	}
	manager.acceptSessions = true
	manager.sessionMutex.Unlock()

	return nil
}

// rejects all new sessions. blocking = true waits for all current session acceptions to finish (after which no new sessions/identities will be added)
func (manager *SessionManager) RejectSessions(blocking bool) error {
	manager.acceptSessionsMutex.Lock()
	defer manager.acceptSessionsMutex.Unlock()

	manager.sessionMutex.Lock()
	if !manager.acceptSessions {
		manager.sessionMutex.Unlock()
		return errors.New("session manager already rejecting sessions")
	}
	manager.acceptSessions = false
	manager.sessionMutex.Unlock()

	if blocking {
		manager.waitgroup.Wait()
	}
	return nil
}

func (manager *SessionManager) GetSession(sessionId string) *Session {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()
	if session, ok := manager.sessions[sessionId]; ok && session != nil {
		return session
	}
	return nil
}

func (manager *SessionManager) GetIdentities() []string {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()

	identities := make([]string, 0, len(manager.identities))
	for identity := range manager.identities {
		identities = append(identities, identity)
	}
	return identities
}

func (manager *SessionManager) GetSessions(identityString string) []*Session {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()

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
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()

	if _, ok := manager.identities[identityString]; ok {
		for _, session := range manager.identities[identityString].sessions {
			if session != nil {
				return true
			}
		}
	}
	return false
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
