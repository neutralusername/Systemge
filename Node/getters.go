package Node

import (
	"Systemge/Utilities"
)

func (node *Node) GetName() string {
	return node.config.Name
}

func (node *Node) GetLogger() *Utilities.Logger {
	return node.logger
}
