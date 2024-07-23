package Dashboard

import "Systemge/Node"

type NodeSystemgeCounters struct {
	Name           string `json:"name"`
	IncSyncReq     uint32 `json:"incSyncReq"`
	IncSyncRes     uint32 `json:"incSyncRes"`
	IncAsync       uint32 `json:"incAsync"`
	OutSyncReq     uint32 `json:"outSyncReq"`
	OutSyncRes     uint32 `json:"outSyncRes"`
	OutAsync       uint32 `json:"outAsync"`
	BytesSent      uint64 `json:"bytesSent"`
	BytesRecveived uint64 `json:"bytesReceived"`
}

func newNodeSystemgeCounters(node *Node.Node) NodeSystemgeCounters {
	return NodeSystemgeCounters{
		Name:           node.GetName(),
		IncSyncReq:     node.GetSystemgeIncomingSyncRequestMessageCounter(),
		IncSyncRes:     node.GetSystemgeIncomingSyncResponseMessageCounter(),
		IncAsync:       node.GetSystemgeIncomingAsyncMessageCounter(),
		OutSyncReq:     node.GetSystemgeOutgoingSyncRequestMessageCounter(),
		OutSyncRes:     node.GetSystemgeOutgoingSyncResponseMessageCounter(),
		OutAsync:       node.GetSystemgeOutgoingAsyncMessageCounter(),
		BytesSent:      node.GetSystemgeBytesSentCounter(),
		BytesRecveived: node.GetSystemgeBytesReceivedCounter(),
	}
}
