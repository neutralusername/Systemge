package Oauth2Server

import (
	"time"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *Server) GetSession(sessionId string) *session {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.sessions[sessionId]
}

func (server *Server) getSessionForIdentity(identity string, keyValuePairs map[string]interface{}) *session {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	session := server.identities[identity]
	if session == nil {
		session = server.createSession(identity, keyValuePairs)
		if infoLogger := server.infoLogger; infoLogger != nil {
			infoLogger.Log(Event.New("Created session \""+session.sessionId+"\" with identity \""+session.identity+"\"", nil).Error())
		}
	} else {
		session.watchdog.Reset(time.Duration(server.config.SessionLifetimeMs) * time.Millisecond)
		session.expired = false
		if infoLogger := server.infoLogger; infoLogger != nil {
			infoLogger.Log(Event.New("Refreshed session \""+session.sessionId+"\" with identity \""+session.identity+"\"", nil).Error())
		}
	}
	return session
}

func (server *Server) createSession(identity string, keyValuePairs map[string]interface{}) (session *session) {
	for {
		sessionId := server.randomizer.GenerateRandomString(32, Tools.ALPHA_NUMERIC)
		if _, ok := server.sessions[sessionId]; !ok {
			session = newSession(sessionId, identity, keyValuePairs)
			server.sessions[sessionId] = session
			server.identities[identity] = session
			session.watchdog = time.AfterFunc(time.Duration(server.config.SessionLifetimeMs)*time.Millisecond, server.getRemoveSessionFunc(session))
			break
		}
	}
	return
}

func (server *Server) getRemoveSessionFunc(session *session) func() {
	return func() {
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
		delete(server.sessions, session.sessionId)
		delete(server.identities, session.identity)
		if infoLogger := server.infoLogger; infoLogger != nil {
			infoLogger.Log(Event.New("Removed session \""+session.sessionId+"\" with identity \""+session.identity+"\"", nil).Error())
		}
	}
}

func (server *Server) handleSessionRequests() {
	for sessionRequest := range server.sessionRequestChannel {
		if sessionRequest == nil {
			return
		}
		if infoLogger := server.infoLogger; infoLogger != nil {
			infoLogger.Log(Event.New("Handling session request with access token \""+sessionRequest.token.AccessToken+"\"", nil).Error())
		}
		server.handleSessionRequest(sessionRequest)
	}
}

func (server *Server) handleSessionRequest(sessionRequest *oauth2SessionRequest) {
	identity, keyValuePairs, err := server.config.TokenHandler(server.config.OAuth2Config, sessionRequest.token)
	if err != nil {
		sessionRequest.sessionChannel <- nil
		if warningLogger := server.warningLogger; warningLogger != nil {
			warningLogger.Log(Event.New("Failed handling session request for access token \""+sessionRequest.token.AccessToken+"\"", err).Error())
		}
		return
	}
	if identity == "" {
		sessionRequest.sessionChannel <- nil
		if warningLogger := server.warningLogger; warningLogger != nil {
			warningLogger.Log(Event.New("No session identity for access token \""+sessionRequest.token.AccessToken+"\"", nil).Error())
		}
		return
	}
	sessionRequest.sessionChannel <- server.getSessionForIdentity(identity, keyValuePairs)
}
