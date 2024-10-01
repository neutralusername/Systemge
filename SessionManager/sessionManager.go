package SessionManager

import (
	"math"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
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
		config.SessionIdAlphabet = Tools.ALPHA_NUMERIC
	}
	return &SessionManager{
		name:   name,
		config: config,

		instanceId: Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

		sessions:   make(map[string]*Session),
		identities: make(map[string]*Identity),

		isStarted: false,

		maxTotalSessions: math.Pow(float64(len(config.SessionIdAlphabet)), float64(config.SessionIdLength)) * 0.8,

		eventHandler: eventHandler,
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

func (manager *SessionManager) GetSessions(identityString string) []*Session {
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
