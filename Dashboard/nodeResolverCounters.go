package Dashboard

import (
	"github.com/neutralusername/Systemge/Node"
)

type NodeResolverCounters struct {
	Name               string `json:"name"`
	BytesReceived      uint64 `json:"bytesReceived"`
	BytesSent          uint64 `json:"bytesSent"`
	ConfigRequests     uint32 `json:"configRequests"`
	ResolutionRequests uint32 `json:"resolutionRequests"`
}

func newNodeResolverCounters(node *Node.Node) NodeResolverCounters {
	return NodeResolverCounters{
		Name:               node.GetName(),
		BytesReceived:      node.RetrieveResolverBytesReceivedCounter(),
		BytesSent:          node.RetrieveResolverBytesSentCounter(),
		ConfigRequests:     node.RetrieveResolverConfigRequestCounter(),
		ResolutionRequests: node.RetrieveResolverResolutionRequestCounter(),
	}
}
