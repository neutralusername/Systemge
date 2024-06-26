package Module

import "Systemge/Node"

// equivalent to Node.New
func NewNode(config *Node.Config, application Node.Application, httpComponent Node.HTTPComponent, websocketComponent Node.WebsocketComponent) *Node.Node {
	return Node.New(config, application, httpComponent, websocketComponent)
}
