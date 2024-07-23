package Dashboard

import "Systemge/Node"

type NodeSystemgeCounters struct {
	Name       string `json:"name"`
	IncSyncReq uint32 `json:"incSyncReq"`
	IncSyncRes uint32 `json:"incSyncRes"`
	IncAsync   uint32 `json:"incAsync"`
	OutSyncReq uint32 `json:"outSyncReq"`
	OutSyncRes uint32 `json:"outSyncRes"`
	OutAsync   uint32 `json:"outAsync"`
}

func newNodeSystemgeCounters(node *Node.Node) NodeSystemgeCounters {
	incSyncReq, incSyncRes, incAsync, outSyncReq, outSyncRes, outAsync := node.GetSystemgeCounters()
	return NodeSystemgeCounters{
		Name:       node.GetName(),
		IncSyncReq: incSyncReq,
		IncSyncRes: incSyncRes,
		IncAsync:   incAsync,
		OutSyncReq: outSyncReq,
		OutSyncRes: outSyncRes,
		OutAsync:   outAsync,
	}
}
