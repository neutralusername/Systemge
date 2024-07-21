package Dashboard

import (
	"Systemge/Node"
	"encoding/json"
)

type NodeStatus struct {
	Name   string `json:"name"`
	Status bool   `json:"status"`
}

type NodeCommands struct {
	Name     string   `json:"name"`
	Commands []string `json:"commands"`
}

func newNodeStatus(node *Node.Node) NodeStatus {
	return NodeStatus{
		Name:   node.GetName(),
		Status: node.IsStarted(),
	}
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

func jsonMarshal(data interface{}) string {
	json, _ := json.Marshal(data)
	return string(json)
}
