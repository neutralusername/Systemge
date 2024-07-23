package Dashboard

import "Systemge/Node"

type NodeWebsocketCounters struct {
	Name        string `json:"name"`
	ClientCount int    `json:"clientCount"`
	GroupCount  int    `json:"groupCount"`
	Inc         uint32 `json:"inc"`
	Out         uint32 `json:"out"`
}

func newNodeWebsocketCounters(node *Node.Node) NodeWebsocketCounters {
	return NodeWebsocketCounters{
		Name:        node.GetName(),
		ClientCount: node.GetWebsocketClientCount(),
		GroupCount:  node.GetWebsocketGroupCount(),
		Inc:         node.GetWebsocketIncomingMessageCounter(),
		Out:         node.GetWebsocketOutgoingMessageCounter(),
	}
}
