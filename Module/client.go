package Module

import "Systemge/Node"

// equivalent to Node.New
func NewNode(config *Node.Config, application Node.Application, httpApplication Node.HTTPComponent, websocketApplication Node.WebsocketComponent) *Node.Node {
	return Node.New(config, application, httpApplication, websocketApplication)
}
