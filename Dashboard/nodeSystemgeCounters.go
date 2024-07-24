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
		IncSyncReq:     node.RetrieveSystemgeIncomingSyncRequestMessageCounter(),
		IncSyncRes:     node.RetrieveSystemgeIncomingSyncResponseMessageCounter(),
		IncAsync:       node.RetrieveSystemgeIncomingAsyncMessageCounter(),
		OutSyncReq:     node.RetrieveSystemgeOutgoingSyncRequestMessageCounter(),
		OutSyncRes:     node.RetrieveSystemgeOutgoingSyncResponseMessageCounter(),
		OutAsync:       node.RetrieveSystemgeOutgoingAsyncMessageCounter(),
		BytesSent:      node.RetrieveSystemgeBytesSentCounter(),
		BytesRecveived: node.RetrieveSystemgeBytesReceivedCounter(),
	}
}
