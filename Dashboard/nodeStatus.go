package Dashboard

import (
	"github.com/neutralusername/Systemge/Node"
)

type NodeStatus struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
}

func newNodeStatus(node *Node.Node) NodeStatus {
	return NodeStatus{
		Name:   node.GetName(),
		Status: node.GetStatus(),
	}
}
