package Http

import (
	"net"
	"net/http"
)

func (server *Server) accessControllWrapper(handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			Send403(w, r)
			return
		}
		if server.blacklist != nil {
			if server.blacklist.Contains(ip) {
				Send403(w, r)
				return
			}
		}
		if server.whitelist != nil && server.whitelist.ElementCount() > 0 {
			if !server.whitelist.Contains(ip) {
				Send403(w, r)
				return
			}
		}
		handler(w, r)
	}
}
