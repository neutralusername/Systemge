package Spawner

import (
	"Systemge/Error"
	"Systemge/Node"
)

func (spawner *Spawner) GetCommandHandlers() map[string]Node.CommandHandler {
	return map[string]Node.CommandHandler{
		"spawnedNodes": func(node *Node.Node, args []string) (string, error) {
			spawner.mutex.Lock()
			defer spawner.mutex.Unlock()
			resultStr := ""
			for _, spawnedNode := range spawner.spawnedNodes {
				resultStr += spawnedNode.GetName() + ";"
			}
			return resultStr, nil
		},
		"spawnNode": func(node *Node.Node, args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			spawner.mutex.Lock()
			defer spawner.mutex.Unlock()
			err := spawner.StartNode(node, args[0])
			if err != nil {
				return "", Error.New("Error starting node "+args[0], err)
			}
			return "success", nil
		},
		"endNode": func(node *Node.Node, args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			spawner.mutex.Lock()
			defer spawner.mutex.Unlock()
			err := spawner.EndNode(node, args[0])
			if err != nil {
				return "", Error.New("Error ending node "+args[0], err)
			}
			return "success", nil
		},
	}
}
