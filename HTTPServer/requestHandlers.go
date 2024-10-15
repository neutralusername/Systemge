package HTTPServer

import (
	"net/http"

	"golang.org/x/oauth2"
)

func Oauth2AuthCallback(oauth2Config *oauth2.Config, oauth2State string, tokenHandler func(*oauth2.Token, http.ResponseWriter)) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		state := httpRequest.FormValue("state")
		if state != oauth2State {
			Send403(responseWriter, httpRequest)
			return
		}
		code := httpRequest.FormValue("code")
		token, err := oauth2Config.Exchange(httpRequest.Context(), code)
		if err != nil {
			Send403(responseWriter, httpRequest)
			return
		}
		tokenHandler(token, responseWriter)
	}
}

func Oauth2AuthRedirect(oauth2Config *oauth2.Config, oauth2State string, redirectUrl string) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		url := redirectUrl
		if url == "" {
			url = oauth2Config.AuthCodeURL(oauth2State)
		}
		http.Redirect(responseWriter, httpRequest, url, http.StatusTemporaryRedirect)
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
