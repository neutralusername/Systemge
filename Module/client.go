package Module

import (
	"Systemge/Client"
)

// equivalent to Client.New
func NewClient(config *Client.Config, application Client.Application, httpApplication Client.HTTPComponent, websocketApplication Client.WebsocketComponent) *Client.Client {
	return Client.New(config, application, httpApplication, websocketApplication)
}
