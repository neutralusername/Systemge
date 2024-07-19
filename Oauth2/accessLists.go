package Oauth2

import (
	"Systemge/Error"
	"net"
)

func (server *Server) validateAddress(address string) error {
	ipAddress, _, err := net.SplitHostPort(address)
	if err != nil {
		return Error.New("Failed to get IP address and port from \""+address+"\"", err)
	}
	if server.isBlacklisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to blacklist", nil)
	}
	if !server.isWhitelisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to whitelist", nil)
	}
	return nil
}

func (server *Server) isBlacklisted(ip string) bool {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.blacklist[ip]
}

func (server *Server) isWhitelisted(ip string) bool {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if len(server.whitelist) == 0 {
		return true
	}
	return server.whitelist[ip]
}

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
