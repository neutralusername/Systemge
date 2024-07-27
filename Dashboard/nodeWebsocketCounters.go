package Dashboard

import "github.com/neutralusername/Systemge/Node"

type NodeWebsocketCounters struct {
	Name          string `json:"name"`
	ClientCount   int    `json:"clientCount"`
	GroupCount    int    `json:"groupCount"`
	Inc           uint32 `json:"inc"`
	Out           uint32 `json:"out"`
	BytesSent     uint64 `json:"bytesSent"`
	BytesReceived uint64 `json:"bytesReceived"`
}

func newNodeWebsocketCounters(node *Node.Node) NodeWebsocketCounters {
	return NodeWebsocketCounters{
		Name:          node.GetName(),
		ClientCount:   node.GetWebsocketClientCount(),
		GroupCount:    node.GetWebsocketGroupCount(),
		Inc:           node.RetrieveWebsocketIncomingMessageCounter(),
		Out:           node.RetrieveWebsocketOutgoingMessageCounter(),
		BytesSent:     node.RetrieveWebsocketBytesSentCounter(),
		BytesReceived: node.RetrieveWebsocketBytesReceivedCounter(),
	}
}
