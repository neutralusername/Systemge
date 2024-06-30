package Spawner

import (
	"Systemge/Error"
	"Systemge/Node"
)

func (spawner *Spawner) GetCustomCommandHandlers() map[string]Node.CustomCommandHandler {
	return map[string]Node.CustomCommandHandler{
		"spawnedNodes": func(node *Node.Node, args []string) error {
			spawner.mutex.Lock()
			defer spawner.mutex.Unlock()
			for _, spawnedNode := range spawner.spawnedNodes {
				println(spawnedNode.GetName())
			}
			return nil
		},
		"spawnNode": func(node *Node.Node, args []string) error {
			if len(args) != 1 {
				return Error.New("Invalid number of arguments", nil)
			}
			spawner.mutex.Lock()
			defer spawner.mutex.Unlock()
			err := spawner.StartNode(node, args[0])
			if err != nil {
				return Error.New("Error starting node "+args[0], err)
			}
			return nil
		},
		"endNode": func(node *Node.Node, args []string) error {
			if len(args) != 1 {
				return Error.New("Invalid number of arguments", nil)
			}
			spawner.mutex.Lock()
			defer spawner.mutex.Unlock()
			err := spawner.EndNode(node, args[0])
			if err != nil {
				return Error.New("Error ending node "+args[0], err)
			}
			return nil
		},
	}
}
