package Module

import "Systemge/Node"

// equivalent to Client.New
func NewClient(config *Node.Config, application Node.Application, httpApplication Node.HTTPComponent, websocketApplication Node.WebsocketComponent) *Node.Node {
	return Node.New(config, application, httpApplication, websocketApplication)
}
