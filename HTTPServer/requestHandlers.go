package HTTPServer

import (
	"Systemge/Application"
	"net/http"
)

func SendDirectory(path string) Application.HTTPRequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		http.FileServer(http.Dir(path)).ServeHTTP(responseWriter, httpRequest)
	}
}

func RedirectTo(toURL string) Application.HTTPRequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		http.Redirect(responseWriter, httpRequest, toURL, http.StatusMovedPermanently)
	}
}

func SendHTTPResponseFull(statusCode int, headerKeyValuePairs map[string]string, body string) Application.HTTPRequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		for key, value := range headerKeyValuePairs {
			responseWriter.Header().Set(key, value)
		}
		responseWriter.WriteHeader(statusCode)
		responseWriter.Write([]byte(body))
	}
}

func SendHTTPResponseCodeAndBody(statusCode int, body string) Application.HTTPRequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		responseWriter.WriteHeader(statusCode)
		responseWriter.Write([]byte(body))
	}
}

func SendHTTPResponseCode(statusCode int) Application.HTTPRequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		responseWriter.WriteHeader(statusCode)
	}
}
