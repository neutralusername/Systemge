package Oauth2

func (server *Server) GetSession(sessionId string) *Session {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.sessions[sessionId]
}
