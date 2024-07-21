package Dashboard

import (
	"Systemge/Node"
)

type NodeStatus struct {
	Name   string `json:"name"`
	Status bool   `json:"status"`
}

func newNodeStatus(node *Node.Node) NodeStatus {
	return NodeStatus{
		Name:   node.GetName(),
		Status: node.IsStarted(),
	}
}
