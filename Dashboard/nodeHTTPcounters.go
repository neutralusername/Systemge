package Dashboard

import "github.com/neutralusername/Systemge/Node"

type NodeHTTPCounters struct {
	HTTPRequestCount uint64 `json:"requestCount"`
	Name             string `json:"name"`
}

func newHTTPCounters(node *Node.Node) *NodeHTTPCounters {
	return &NodeHTTPCounters{
		HTTPRequestCount: node.GetHTTPRequestCounter(),
		Name:             node.GetName(),
	}
}
