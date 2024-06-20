package WebsocketServer

import (
	"Systemge/Application"
)

type HandshakeApplication struct {
	websocketServer *Server
	pattern         string
}

func (h *HandshakeApplication) GetHTTPRequestHandlers() map[string]Application.HTTPRequestHandler {
	return map[string]Application.HTTPRequestHandler{
		h.pattern: PromoteToWebsocket(h.websocketServer),
	}
}

func NewHandshakeApplication(pattern string, websocketServer *Server) Application.HTTPApplication {
	return &HandshakeApplication{
		websocketServer: websocketServer,
		pattern:         pattern,
	}
}
