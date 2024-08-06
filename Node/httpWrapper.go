package Node

import (
	"net"
	"net/http"

	"github.com/neutralusername/Systemge/HTTP"
)

func (httpComponent *httpComponent) httpRequestWrapper(handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpComponent.requestCounter.Add(1)
		r.Body = http.MaxBytesReader(w, r.Body, httpComponent.config.MaxBodyBytes)

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			HTTP.Send403(w, r)
			return
		}
		if httpComponent.server.GetBlacklist() != nil {
			if httpComponent.server.GetBlacklist().Contains(ip) {
				HTTP.Send403(w, r)
				return
			}
		}
		if httpComponent.server.GetWhitelist() != nil && httpComponent.server.GetWhitelist().ElementCount() > 0 {
			if !httpComponent.server.GetWhitelist().Contains(ip) {
				HTTP.Send403(w, r)
				return
			}
		}
		handler(w, r)
	}
}
