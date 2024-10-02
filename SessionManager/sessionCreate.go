package SessionManager

import (
	"errors"

	"github.com/neutralusername/Systemge/Tools"
)

func (manager *Manager) CreateSession(identityString string, keyValuePairs map[string]any) (*Session, error) {

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

	sessionId := Tools.GenerateRandomString(manager.config.SessionIdLength, manager.config.SessionIdAlphabet)
	for {
		if _, ok := manager.sessions[sessionId]; !ok {
			break
		}
		sessionId = Tools.GenerateRandomString(manager.config.SessionIdLength, manager.config.SessionIdAlphabet)
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
	session.timeout = Tools.NewTimeout(
		manager.config.SessionLifetimeMs,
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
func (manager *Manager) cleanupSession(session *Session) {
	manager.sessionMutex.Lock()
	defer manager.sessionMutex.Unlock()
	delete(manager.sessions, session.GetId())
	delete(manager.identities[session.GetIdentity()].sessions, session.GetId())
	if len(manager.identities[session.GetIdentity()].sessions) == 0 {
		delete(manager.identities, session.GetIdentity())
	}
}
