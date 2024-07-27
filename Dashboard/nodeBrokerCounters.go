package Dashboard

import (
	"github.com/neutralusername/Systemge/Broker"
	"github.com/neutralusername/Systemge/Node"
)

type NodeBrokerCounters struct {
	Name             string `json:"name"`
	BytesReceived    uint64 `json:"bytesReceived"`
	BytesSent        uint64 `json:"bytesSent"`
	IncomingMessages uint32 `json:"incomingMessages"`
	OutgoingMessages uint32 `json:"outgoingMessages"`
	ConfigRequests   uint32 `json:"configRequests"`
}

func newNodeBrokerCounters(node *Node.Node) NodeBrokerCounters {
	broker := node.GetApplication().(*Broker.Broker)
	return NodeBrokerCounters{
		Name:             node.GetName(),
		BytesReceived:    broker.RetrieveBytesReceivedCounter(),
		BytesSent:        broker.RetrieveBytesSentCounter(),
		IncomingMessages: broker.RetrieveIncomingMessageCounter(),
		OutgoingMessages: broker.RetrieveOutgoingMessageCounter(),
		ConfigRequests:   broker.RetrieveConfigRequestCounter(),
	}
}
