package Oauth2Server

func (server *Server) handleSessionsCommand(args []string) (string, error) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	returnString := ""
	for _, session := range server.sessions {
		returnString += session.identity + ":" + session.sessionId + ";"
	}
	return returnString, nil
}
