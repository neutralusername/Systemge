package HTTPServer

import (
	"net"
	"net/http"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (server *HTTPServer) httpRequestWrapper(pattern string, handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server.statusMutex.RLock()
		defer server.statusMutex.RUnlock()
		if server.status != Status.Started {
			Send403(w, r)
			return
		}

		if event := server.onEvent(Event.NewInfo(
			Event.HandlingHttpRequest,
			"Handling http request",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:  Event.HttpRequest,
				Event.Pattern:       pattern,
				Event.ClientType:    Event.HttpRequest,
				Event.ClientAddress: r.RemoteAddr,
			},
		)); !event.IsInfo() {
			Send403(w, r)
			return
		}

		server.requestCounter.Add(1)
		r.Body = http.MaxBytesReader(w, r.Body, server.config.MaxBodyBytes)

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.SplittingHostPortFailed,
				err.Error(),
				Event.Context{
					Event.Circumstance:  Event.HttpRequest,
					Event.ClientType:    Event.HttpRequest,
					Event.ClientAddress: r.RemoteAddr,
				}),
			)
			Send403(w, r)
			return
		}
		if server.GetBlacklist() != nil {
			if server.GetBlacklist().Contains(ip) {
				if event := server.onEvent(Event.NewWarning(
					Event.Blacklisted,
					"Client not accepted",
					Event.Cancel,
					Event.Cancel,
					Event.Continue,
					Event.Context{
						Event.Circumstance:  Event.HttpRequest,
						Event.Pattern:       pattern,
						Event.ClientType:    Event.HttpRequest,
						Event.ClientAddress: r.RemoteAddr,
					},
				)); !event.IsWarning() {
					Send403(w, r)
					return
				}
			}
		}
		if server.GetWhitelist() != nil && server.GetWhitelist().ElementCount() > 0 {
			if !server.GetWhitelist().Contains(ip) {
				if event := server.onEvent(Event.NewWarning(
					Event.NotWhitelisted,
					"Client not accepted",
					Event.Cancel,
					Event.Cancel,
					Event.Continue,
					Event.Context{
						Event.Circumstance:  Event.HttpRequest,
						Event.Pattern:       pattern,
						Event.ClientType:    Event.HttpRequest,
						Event.ClientAddress: r.RemoteAddr,
					},
				)); !event.IsWarning() {
					Send403(w, r)
					return
				}
			}
		}

		handler(w, r)

		server.onEvent(Event.NewInfoNoOption(
			Event.HandledHttpRequest,
			"Handled http request",
			Event.Context{
				Event.Circumstance:  Event.HttpRequest,
				Event.Pattern:       pattern,
				Event.ClientType:    Event.HttpRequest,
				Event.ClientAddress: r.RemoteAddr,
			},
		))
	}
}

func SendDirectory(path string) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		http.FileServer(http.Dir(path)).ServeHTTP(responseWriter, httpRequest)
	}
}

func RedirectTo(toURL string) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		http.Redirect(responseWriter, httpRequest, toURL, http.StatusMovedPermanently)
	}
}

func SendHTTPResponseFull(statusCode int, headerKeyValuePairs map[string]string, body string) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		for key, value := range headerKeyValuePairs {
			responseWriter.Header().Set(key, value)
		}
		responseWriter.WriteHeader(statusCode)
		responseWriter.Write([]byte(body))
	}
}

func SendHTTPResponseCodeAndBody(statusCode int, body string) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		responseWriter.WriteHeader(statusCode)
		responseWriter.Write([]byte(body))
	}
}

func SendHTTPResponseCode(statusCode int) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		responseWriter.WriteHeader(statusCode)
	}
}

func Send404(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	responseWriter.WriteHeader(http.StatusNotFound)
	responseWriter.Write([]byte("404 page not found"))
}

func Send403(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	responseWriter.WriteHeader(http.StatusForbidden)
	responseWriter.Write([]byte("403 forbidden"))
}
