package Dashboard

import "github.com/neutralusername/Systemge/Node"

type AddNode struct {
	Name     string   `json:"name"`
	Status   int      `json:"status"`
	Commands []string `json:"commands"`
}

func newAddNode(node *Node.Node) AddNode {
	commands := []string{}
	for command := range node.GetCommandHandlers() {
		commands = append(commands, command)

	}
	return AddNode{
		Commands: commands,
		Name:     node.GetName(),
		Status:   node.GetStatus(),
	}
}
