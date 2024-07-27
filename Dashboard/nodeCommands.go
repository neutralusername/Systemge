package Dashboard

import "github.com/neutralusername/Systemge/Node"

type NodeCommands struct {
	Name     string   `json:"name"`
	Commands []string `json:"commands"`
}

func newNodeCommands(node *Node.Node) NodeCommands {
	commands := []string{}
	for command := range node.GetCommandHandlers() {
		commands = append(commands, command)

	}
	return NodeCommands{
		Commands: commands,
		Name:     node.GetName(),
	}
}
