package Tools

import (
	"errors"
	"math"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

type SessionManager struct {
	name             string
	config           *Config.SessionManager
	maxTotalSessions float64

	instanceId string
	sessionId  string

	sessionMutex sync.RWMutex
	identities   map[string]*Identity
	sessions     map[string]*Session

	isStarted   bool
	waitgroup   sync.WaitGroup
	statusMutex sync.Mutex

	eventHandler Event.Handler
}

func NewSessionManager(name string, config *Config.SessionManager, eventHandler Event.Handler) *SessionManager {
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
		name:   name,
		config: config,

		instanceId: GenerateRandomString(Constants.InstanceIdLength, ALPHA_NUMERIC),

		sessions:   make(map[string]*Session),
		identities: make(map[string]*Identity),

		isStarted: false,

		maxTotalSessions: math.Pow(float64(len(config.SessionIdAlphabet)), float64(config.SessionIdLength)) * 0.8,

		eventHandler: eventHandler,
	}
}

func (manager *SessionManager) CreateSession(identityString string) (*Session, error) {
	manager.sessionMutex.Lock()
	if !manager.isStarted {
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
		manager.cleanupSession(session)
		return nil, err
	} else {
		manager.sessionMutex.Lock()
		manager.sessions[session.GetId()] = session
		identity.sessions[session.GetId()] = session
		manager.sessionMutex.Unlock()
	}

	session.timeout = NewTimeout(
		manager.config.SessionLifetimeMs,
		func() {
			manager.cleanupSession(session)

			manager.onExpire(session)
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

func (manager *SessionManager) Start() error {
	manager.statusMutex.Lock()
	defer manager.statusMutex.Unlock()

	if event := manager.onEvent(Event.NewInfo(
		Event.ServiceStarting,
		"starting sessionManager",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStart,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	manager.sessionMutex.Lock()
	if manager.isStarted {
		manager.sessionMutex.Unlock()
		manager.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStarted,
			"session manager already started",
			Event.Context{
				Event.Circumstance: Event.ServiceStart,
			},
		))
		return errors.New("session manager already accepting sessions")
	}
	manager.sessionId = GenerateRandomString(manager.config.SessionIdLength, manager.config.SessionIdAlphabet)
	manager.isStarted = true
	manager.sessionMutex.Unlock()

	manager.onEvent(Event.NewInfoNoOption(
		Event.ServiceStarted,
		"sessionManager started",
		Event.Context{
			Event.Circumstance: Event.ServiceStart,
		},
	))
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
		for _, session := range manager.identities[identity].sessions {
			if session != nil {
				identities = append(identities, identity)
				break
			}
		}
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
		if len(sessions) > 0 {
			return sessions
		}
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

func (manager *SessionManager) GetStatus() int {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()
	if manager.isStarted {
		return Status.Started
	}
	return Status.Stopped
}

func (server *SessionManager) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(server.GetContext())
	if server.eventHandler != nil {
		server.eventHandler(event)
	}
	return event
}
func (server *SessionManager) GetContext() Event.Context {
	return Event.Context{
		Event.ServiceType:   Event.SessionManager,
		Event.ServiceName:   server.name,
		Event.ServiceStatus: Status.ToString(server.GetStatus()),
		Event.Function:      Event.GetCallerFuncName(2),
		Event.InstanceId:    server.instanceId,
		Event.SessionId:     server.sessionId,
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
