package SessionManager

import (
	"math"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type Manager struct {
	name             string
	config           *Config.SessionManager
	maxTotalSessions float64

	instanceId string
	sessionId  string

	sessionMutex sync.RWMutex
	identities   map[string]*Identity
	sessions     map[string]*Session

	onCreateSession func(*Session) error
	onRemoveSession func(*Session)

	isStarted   bool
	waitgroup   sync.WaitGroup
	statusMutex sync.Mutex
}

func New(name string, config *Config.SessionManager, onCreateSession func(*Session) error, onRemoveSession func(*Session)) *Manager {
	if config == nil {
		config = &Config.SessionManager{}
	}
	if config.SessionIdLength == 0 {
		config.SessionIdLength = 32
	}
	if config.SessionIdAlphabet == "" {
		config.SessionIdAlphabet = Tools.ALPHA_NUMERIC
	}
	return &Manager{
		name:   name,
		config: config,

		instanceId: Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

		sessions:   make(map[string]*Session),
		identities: make(map[string]*Identity),

		isStarted: false,

		onCreateSession: onCreateSession,
		onRemoveSession: onRemoveSession,

		maxTotalSessions: math.Pow(float64(len(config.SessionIdAlphabet)), float64(config.SessionIdLength)) * 0.8,
	}
}

func (manager *Manager) GetSession(sessionId string) *Session {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()
	if session, ok := manager.sessions[sessionId]; ok {
		return session
	}
	return nil
}

func (manager *Manager) GetIdentities() []string {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()

	identities := make([]string, 0, len(manager.identities))
	for identity := range manager.identities {
		identities = append(identities, identity)
	}
	return identities
}

func (manager *Manager) GetSessions(identityString string) []*Session {
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

func (manager *Manager) IdentityExists(identityString string) bool {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()

	if _, ok := manager.identities[identityString]; ok {
		return true
	}
	return false
}

func (manager *Manager) GetStatus() int {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()
	if manager.isStarted {
		return Status.Started
	}
	return Status.Stopped
}
