package Spawner

import "Systemge/Node"

func (spawner *Spawner) GetCustomCommandHandlers() map[string]Node.CustomCommandHandler {
	return map[string]Node.CustomCommandHandler{
		"command": func(node *Node.Node, args []string) error {
			println("executing command")
			return nil
		},
	}
}
