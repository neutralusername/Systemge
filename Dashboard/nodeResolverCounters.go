package Dashboard

import (
	"Systemge/Node"
	"Systemge/Resolver"
)

type NodeResolverCounters struct {
	Name               string `json:"name"`
	BytesReceived      uint64 `json:"bytesReceived"`
	BytesSent          uint64 `json:"bytesSent"`
	ConfigRequests     uint32 `json:"configRequests"`
	ResolutionRequests uint32 `json:"resolutionRequests"`
}

func newNodeResolverCounters(node *Node.Node) NodeResolverCounters {
	resolver := node.GetApplication().(*Resolver.Resolver)
	return NodeResolverCounters{
		Name:               node.GetName(),
		BytesReceived:      resolver.RetrieveBytesReceivedCounter(),
		BytesSent:          resolver.RetrieveBytesSentCounter(),
		ConfigRequests:     resolver.RetrieveConfigRequestCounter(),
		ResolutionRequests: resolver.RetrieveResolutionRequestCounter(),
	}
}
