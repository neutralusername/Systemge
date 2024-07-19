package Oauth2

func (server *Server) addToBlacklist(ips ...string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, ip := range ips {
		server.blacklist[ip] = true
	}
}

func (server *Server) addToWhitelist(ips ...string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, ip := range ips {
		server.whitelist[ip] = true
	}
}

func (server *Server) removeFromBlacklist(ips ...string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, ip := range ips {
		delete(server.blacklist, ip)
	}
}

func (server *Server) removeFromWhitelist(ips ...string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, ip := range ips {
		delete(server.whitelist, ip)
	}
}
