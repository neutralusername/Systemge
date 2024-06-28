package Module

import (
	"Systemge/Config"
	"Systemge/Node"
)

// equivalent to Node.New
func NewNode(config Config.Node, application Node.Application, httpComponent Node.HTTPComponent, websocketComponent Node.WebsocketComponent) *Node.Node {
	return Node.New(config, application, httpComponent, websocketComponent)
}
