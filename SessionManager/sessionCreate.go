package SessionManager

import (
	"errors"

	"github.com/neutralusername/Systemge/Tools"
)

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

	session.timeout = Tools.NewTimeout(
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
