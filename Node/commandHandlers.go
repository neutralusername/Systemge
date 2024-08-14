package Node

import (
	"github.com/neutralusername/Systemge/Error"
)

// returns a map of command handlers for the command-line interface
func (node *Node) GetCommandHandlers() map[string]CommandHandler {
	handlers := map[string]CommandHandler{
		"start": func(node *Node, args []string) (string, error) {
			err := node.Start()
			if err != nil {
				return "", Error.New("Failed to start node \""+node.GetName()+"\": "+err.Error(), nil)
			}
			return "success", nil
		},
		"stop": func(node *Node, args []string) (string, error) {
			err := node.Stop()
			if err != nil {
				return "", Error.New("Failed to stop node \""+node.GetName()+"\": "+err.Error(), nil)
			}
			return "success", nil
		},
	}

	if commandHandlerComponent := node.GetCommandHandlerComponent(); commandHandlerComponent != nil {
		commandHandlers := commandHandlerComponent.GetCommandHandlers()
		for command, commandHandler := range commandHandlers {
			handlers[command] = commandHandler
		}
	}
	return handlers
}
