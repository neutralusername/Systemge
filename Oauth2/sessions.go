package Oauth2

import (
	"Systemge/Error"
	"Systemge/Utilities"
	"time"
)

func (server *Server) GetSession(sessionId string) *session {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.sessions[sessionId]
}

func (server *Server) addSession(identity string, keyValuePairs map[string]interface{}) (session *session, err error) {
	sessionId := ""
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.identities[identity] != nil {
		session.watchdog.Reset(time.Duration(server.config.SessionLifetimeMs) * time.Millisecond)
		session.expired = false
		server.config.Logger.Info(Error.New("refreshed session \""+server.identities[identity].sessionId+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
		return server.identities[identity], nil
	}
	for {
		sessionId = server.config.Randomizer.GenerateRandomString(32, Utilities.ALPHA_NUMERIC)
		if _, ok := server.sessions[sessionId]; !ok {
			session = newSession(sessionId, identity, keyValuePairs)
			server.sessions[sessionId] = session
			server.identities[identity] = session
			session.watchdog = time.AfterFunc(time.Duration(server.config.SessionLifetimeMs)*time.Millisecond, func() {
				session.expired = true
				server.mutex.Lock()
				defer server.mutex.Unlock()
				if !session.expired {
					return
				}
				if session.watchdog == nil {
					return
				}
				session.watchdog = nil
				delete(server.sessions, sessionId)
				delete(server.identities, identity)
				server.config.Logger.Info(Error.New("removed session \""+sessionId+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
			})
			break
		}
	}
	server.config.Logger.Info(Error.New("added session \""+sessionId+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
	return session, err
}
