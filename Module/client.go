package Module

import (
	"Systemge/Client"
)

// ClientConfig is a struct that holds the configuration for a client.
// HTTPCert, HTTPKey, WebsocketCert, and WebsocketKey are paths to the certificate and key files which can be left empty if the server does not use TLS.
// WebsocketPattern is the pattern that the websocket server will listen to for websocket connections.

// equivalent to Client.New
func NewClient(config *Client.Config, application Client.Application, httpApplication Client.HTTPApplication, websocketApplication Client.WebsocketApplication) *Client.Client {
	return Client.New(config, application, httpApplication, websocketApplication)
}
