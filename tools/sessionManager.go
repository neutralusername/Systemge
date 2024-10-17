package tools

import (
	"errors"
	"math"
	"sync"

	"github.com/neutralusername/systemge/Config"
	"github.com/neutralusername/systemge/constants"
	"github.com/neutralusername/systemge/status"
)

type SessionManager struct {
	config *Config.SessionManager

	instanceId string
	sessionId  string

	maxTotalSessions float64

	sessionMutex sync.RWMutex
	identities   map[string]*Identity
	sessions     map[string]*Session

	onCreateSession func(*Session) error
	onRemoveSession func(*Session)

	isStarted   bool
	waitgroup   sync.WaitGroup
	statusMutex sync.Mutex
}

type Session struct {
	id            string
	identity      *Identity
	mutex         sync.RWMutex
	keyValuePairs map[string]any

	accepted bool

	timeout *Timeout
}

type Identity struct {
	id       string
	sessions map[string]*Session
}

func NewSessionManager(config *Config.SessionManager, onCreateSession func(*Session) error, onRemoveSession func(*Session)) *SessionManager {
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

		instanceId: GenerateRandomString(constants.InstanceIdLength, ALPHA_NUMERIC),

		sessions:   make(map[string]*Session),
		identities: make(map[string]*Identity),

		isStarted: false,

		onCreateSession: onCreateSession,
		onRemoveSession: onRemoveSession,

		maxTotalSessions: math.Pow(float64(len(config.SessionIdAlphabet)), float64(config.SessionIdLength)) * 0.9,
	}
}

func (manager *SessionManager) Start() error {
	manager.statusMutex.Lock()
	defer manager.statusMutex.Unlock()

	manager.sessionMutex.Lock()
	if manager.isStarted {
		manager.sessionMutex.Unlock()
		return errors.New("session manager already accepting sessions")
	}

	manager.sessionId = GenerateRandomString(manager.config.SessionIdLength, manager.config.SessionIdAlphabet)
	manager.isStarted = true
	manager.sessionMutex.Unlock()

	return nil
}

func (manager *SessionManager) Stop() error {
	manager.statusMutex.Lock()
	defer manager.statusMutex.Unlock()

	manager.sessionMutex.Lock()
	if !manager.isStarted {
		manager.sessionMutex.Unlock()
		return errors.New("session manager already rejecting sessions")
	}

	manager.isStarted = false
	manager.sessionMutex.Unlock()
	manager.waitgroup.Wait()

	return nil
}

func (manager *SessionManager) CreateSession(identityString string, keyValuePairs map[string]any) (*Session, error) {

	if manager.config.MinIdentityLength > 0 && uint32(len(identityString)) < manager.config.MinIdentityLength {
		return nil, errors.New("identity too short")
	}

	if manager.config.MaxIdentityLength > 0 && uint32(len(identityString)) > manager.config.MaxIdentityLength {
		return nil, errors.New("identity too long")
	}

	manager.sessionMutex.Lock()
	if !manager.isStarted {
		manager.sessionMutex.Unlock()
		return nil, errors.New("session manager not accepting sessions")
	}

	manager.waitgroup.Add(1)
	defer manager.waitgroup.Done()

	if len(manager.sessions) >= int(manager.maxTotalSessions) {
		manager.sessionMutex.Unlock()
		return nil, errors.New("maximum number of sessions reached")
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
		keyValuePairs: keyValuePairs,
		accepted:      false,
	}
	identity.sessions[session.GetId()] = session
	manager.sessions[session.GetId()] = session
	manager.sessionMutex.Unlock()

	if manager.onCreateSession != nil {
		if err := manager.onCreateSession(session); err != nil {
			manager.cleanupSession(session)
			return nil, err
		}
	}

	session.accepted = true
	session.timeout = NewTimeout(
		manager.config.SessionLifetimeNs,
		func() {
			manager.cleanupSession(session)
			if manager.onRemoveSession != nil {
				manager.onRemoveSession(session)
			}
		},
		false,
	)

	return session, nil
}
func (manager *SessionManager) cleanupSession(session *Session) {
	manager.sessionMutex.Lock()
	defer manager.sessionMutex.Unlock()
	delete(manager.sessions, session.GetId())
	delete(manager.identities[session.GetIdentity()].sessions, session.GetId())
	if len(manager.identities[session.GetIdentity()].sessions) == 0 {
		delete(manager.identities, session.GetIdentity())
	}
}

func (manager *SessionManager) GetSession(sessionId string) *Session {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()
	if session, ok := manager.sessions[sessionId]; ok {
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

func (manager *SessionManager) GetIdentitySessions(identityString string) []*Session {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()

	if identity, ok := manager.identities[identityString]; ok {
		sessions := []*Session{}
		for _, session := range identity.sessions {
			sessions = append(sessions, session)
		}
		return sessions
	}
	return nil
}

func (manager *SessionManager) GetSessions() []*Session {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()

	sessions := []*Session{}
	for _, session := range manager.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (manager *SessionManager) IdentityExists(identityString string) bool {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()

	if _, ok := manager.identities[identityString]; ok {
		return true
	}
	return false
}

func (manager *SessionManager) GetStatus() int {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()
	if manager.isStarted {
		return status.Started
	}
	return status.Stopped
}

func (identity *Identity) GetId() string {
	return identity.id
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

func (session *Session) IsAccepted() bool {
	return session.accepted
}
