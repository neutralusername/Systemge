package SessionManager

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
)

func (manager *Manager) CreateSession(identityString string, keyValuePairs map[string]any) (*Session, error) {

	if manager.config.MaxIdentityLength > 0 && uint32(len(identityString)) > manager.config.MaxIdentityLength {
		if event := manager.onEvent(Event.NewWarning(
			Event.IdentityTooLong,
			"identity too long",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.SessionCreate,
				Event.Identity:     identityString,
			},
		)); !event.IsInfo() {
			return nil, errors.New("identity too long")
		}
	}

	manager.sessionMutex.Lock()
	if !manager.isStarted {
		manager.sessionMutex.Unlock()
		manager.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStopped,
			"session manager already stopped",
			Event.Context{
				Event.Circumstance: Event.SessionCreate,
				Event.Identity:     identityString,
			},
		))
		return nil, errors.New("session manager not accepting sessions")
	}

	manager.waitgroup.Add(1)
	defer manager.waitgroup.Done()

	if len(manager.sessions) >= int(manager.maxTotalSessions) {
		manager.sessionMutex.Unlock()
		manager.onEvent(Event.NewWarningNoOption(
			Event.MaxTotalSessionsExceeded,
			"max total sessions exceeded",
			Event.Context{
				Event.Circumstance: Event.SessionCreate,
				Event.Identity:     identityString,
			},
		))
		return nil, errors.New("max total sessions exceeded")
	}

	identity, ok := manager.identities[identityString]
	if ok {
		if manager.config.MaxSessionsPerIdentity > 0 && uint32(len(identity.sessions)) >= manager.config.MaxSessionsPerIdentity {
			manager.sessionMutex.Unlock()
			manager.onEvent(Event.NewWarningNoOption(
				Event.MaxSessionsPerIdentityExceeded,
				"max sessions per identity exceeded",
				Event.Context{
					Event.Circumstance: Event.SessionCreate,
					Event.Identity:     identityString,
				},
			))
			return nil, errors.New("max sessions per identity exceeded")
		}
	} else {
		if manager.config.MaxIdentities > 0 && uint32(len(manager.identities)) >= manager.config.MaxIdentities {
			manager.sessionMutex.Unlock()
			manager.onEvent(Event.NewWarningNoOption(
				Event.MaxIdentitiesExceeded,
				"max identities exceeded",
				Event.Context{
					Event.Circumstance: Event.SessionCreate,
					Event.Identity:     identityString,
				},
			))
			return nil, errors.New("max identities exceeded")
		}

		if event := manager.onEvent(Event.NewInfo(
			Event.CreatingIdentity,
			"creating identity",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.SessionCreate,
				Event.Identity:     identityString,
			},
		)); !event.IsInfo() {
			manager.sessionMutex.Unlock()
			return nil, event.GetError()
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

	if event := manager.onEvent(Event.NewInfo(
		Event.SessionAccepting,
		"accepting session",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.SessionCreate,
			Event.Identity:     identityString,
			Event.SessionId:    sessionId,
		},
	)); !event.IsInfo() {
		manager.cleanupSession(session)
		return nil, event.GetError()
	}

	session.accepted = true
	session.timeout = Tools.NewTimeout(
		manager.config.SessionLifetimeMs,
		func() {
			manager.onEvent(Event.NewInfoNoOption(
				Event.SessionDisconnecting,
				"disconnecting session",
				Event.Context{
					Event.Circumstance: Event.SessionDisconnect,
					Event.Identity:     identityString,
					Event.SessionId:    sessionId,
				},
			))

			manager.cleanupSession(session)

			manager.onEvent(Event.NewInfoNoOption(
				Event.SessionDisconnected,
				"session disconnected",
				Event.Context{
					Event.Circumstance: Event.SessionDisconnect,
					Event.Identity:     identityString,
					Event.SessionId:    sessionId,
				},
			))
		},
		false,
	)

	manager.onEvent(Event.NewInfoNoOption(
		Event.SessionAccepted,
		"session created",
		Event.Context{
			Event.Circumstance: Event.SessionCreate,
			Event.Identity:     identityString,
			Event.SessionId:    sessionId,
		},
	))

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
