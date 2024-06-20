package Application

import "net/http"

type HTTPApplication interface {
	// GetHTTPRequestHandlers returns a map of URL patterns to RequestHandlers.
	// The URL pattern is the path of the URL that the RequestHandler will be.
	// HTTP verbs are not considered in the URL pattern and should be handled by the RequestHandler accordingly.
	GetHTTPRequestHandlers() map[string]HTTPRequestHandler
}

// HTTPRequestHandler is a function that handles an HTTP request
type HTTPRequestHandler func(w http.ResponseWriter, r *http.Request)
