package Spawner

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Node"
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
			err := spawner.spawnNode(Config.UnmarshalNewNode(args[0]))
			if err != nil {
				return "", Error.New("Failed spawning node "+args[0], err)
			}
			return "success", nil
		},
		"despawnNode": func(node *Node.Node, args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			spawner.mutex.Lock()
			defer spawner.mutex.Unlock()
			err := spawner.despawnNode(args[0])
			if err != nil {
				return "", Error.New("Failed despawning node "+args[0], err)
			}
			return "success", nil
		},
		"startNode": func(node *Node.Node, args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			spawner.mutex.Lock()
			defer spawner.mutex.Unlock()
			err := spawner.startNode(args[0])
			if err != nil {
				return "", Error.New("Failed starting node "+args[0], err)
			}
			return "success", nil
		},
		"stopNode": func(node *Node.Node, args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			spawner.mutex.Lock()
			defer spawner.mutex.Unlock()
			err := spawner.stopNode(args[0])
			if err != nil {
				return "", Error.New("Failed ending node "+args[0], err)
			}
			return "success", nil
		},
	}
}
