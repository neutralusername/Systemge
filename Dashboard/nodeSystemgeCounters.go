package Dashboard

import "github.com/neutralusername/Systemge/Node"

type NodeSystemgeCounters struct {
	Name          string `json:"name"`
	BytesReceived uint64 `json:"bytesReceived"`
	BytesSent     uint64 `json:"bytesSent"`
}

type NodeSystemgeOutgoingSyncResponsesCounters struct {
	Name                          string `json:"name"`
	OutgoingSyncResponses         uint32 `json:"outgoingSyncResponses"`
	OutgoingSyncSuccessResponses  uint32 `json:"outgoingSyncSuccessResponses"`
	OutgoingSyncFailureResponses  uint32 `json:"outgoingSyncFailureResponses"`
	OutgoingSyncResponseBytesSent uint64 `json:"outgoingSyncResponseBytesSent"`
}

type NodeSystemgeIncomingSyncResponseCounters struct {
	Name                              string `json:"name"`
	IncomingSyncResponses             uint32 `json:"incomingSyncResponses"`
	IncomingSyncSuccessResponses      uint32 `json:"incomingSyncSuccessResponses"`
	IncomingSyncFailureResponses      uint32 `json:"incomingSyncFailureResponses"`
	IncomingSyncResponseBytesReceived uint64 `json:"incomingSyncResponseBytesReceived"`
}

type NodeSystemgeIncomingSyncRequestCounters struct {
	Name                             string `json:"name"`
	IncomingSyncRequests             uint32 `json:"incomingSyncRequests"`
	IncomingSyncRequestBytesReceived uint64 `json:"incomingSyncRequestBytesReceived"`
}

type NOdeSystemgeIncomingAsyncMessageCounters struct {
	Name                              string `json:"name"`
	IncomingAsyncMessages             uint32 `json:"incomingAsyncMessages"`
	IncomingAsyncMessageBytesReceived uint64 `json:"incomingAsyncMessageBytesReceived"`
}

type NodeSystemgeInvalidMessageCounters struct {
	Name                                   string `json:"name"`
	InvalidMessagesFromIncomingConnections uint32 `json:"invalidMessagesFromIncomingConnections"`
	InvalidMessagesFromOutgoingConnections uint32 `json:"invalidMessagesFromOutgoingConnections"`
}

type NodeSystemgeOutgoingSyncRequestCounters struct {
	Name                         string `json:"name"`
	OutgoingSyncRequests         uint32 `json:"outgoingSyncRequests"`
	OutgoingSyncRequestBytesSent uint64 `json:"outgoingSyncRequestBytesSent"`
}

type NodeSystemgeOutgoingAsyncMessageCounters struct {
	Name                          string `json:"name"`
	OutgoingAsyncMessages         uint32 `json:"outgoingAsyncMessages"`
	OutgoingAsyncMessageBytesSent uint64 `json:"outgoingAsyncMessageBytesSent"`
}

type NodeSystemgeIncomingConnectionAttemptsCounters struct {
	Name                                   string `json:"name"`
	IncomingConnectionAttempts             uint32 `json:"incomingConnectionAttempts"`
	IncomingConnectionAttemptsSuccessful   uint32 `json:"incomingConnectionAttemptsSuccessful"`
	IncomingConnectionAttemptsFailed       uint32 `json:"incomingConnectionAttemptsFailed"`
	IncomingConnectionAttemptBytesSent     uint64 `json:"incomingConnectionAttemptBytesSent"`
	IncomingConnectionAttemptBytesReceived uint64 `json:"incomingConnectionAttemptBytesReceived"`
}

type NodeSystemgeOutgoingConnectionAttemptCounters struct {
	Name                                  string `json:"name"`
	OutoingConnectionAttempts             uint32 `json:"outgoingConnectionAttempts"`
	OutoingConnectionAttemptsSuccessful   uint32 `json:"outgoingConnectionAttemptsSuccessful"`
	OutoingConnectionAttemptsFailed       uint32 `json:"outgoingConnectionAttemptsFailed"`
	OutoingConnectionAttemptBytesSent     uint64 `json:"outgoingConnectionAttemptBytesSent"`
	OutoingConnectionAttemptBytesReceived uint64 `json:"outgoingConnectionAttemptBytesReceived"`
}

func newNodeSystemgeCounters(node *Node.Node) NodeSystemgeCounters {
	return NodeSystemgeCounters{
		Name:          node.GetName(),
		BytesReceived: node.RetrieveSystemgeBytesReceivedCounter(),
		BytesSent:     node.RetrieveSystemgeBytesSentCounter(),
	}
}

func newNodeSystemgeIncomingSyncResponseCounters(node *Node.Node) NodeSystemgeIncomingSyncResponseCounters {
	return NodeSystemgeIncomingSyncResponseCounters{
		Name:                              node.GetName(),
		IncomingSyncResponses:             node.RetrieveIncomingSyncResponses(),
		IncomingSyncSuccessResponses:      node.RetrieveIncomingSyncSuccessResponses(),
		IncomingSyncFailureResponses:      node.RetrieveIncomingSyncFailureResponses(),
		IncomingSyncResponseBytesReceived: node.RetrieveIncomingSyncResponseBytesReceived(),
	}
}

func newNodeSystemgeIncomingSyncRequestCounters(node *Node.Node) NodeSystemgeIncomingSyncRequestCounters {
	return NodeSystemgeIncomingSyncRequestCounters{
		Name:                             node.GetName(),
		IncomingSyncRequests:             node.RetrieveIncomingSyncRequests(),
		IncomingSyncRequestBytesReceived: node.RetrieveIncomingSyncRequestBytesReceived(),
	}
}

func newNodeSystemgeIncomingAsyncMessageCounters(node *Node.Node) NOdeSystemgeIncomingAsyncMessageCounters {
	return NOdeSystemgeIncomingAsyncMessageCounters{
		Name:                              node.GetName(),
		IncomingAsyncMessages:             node.RetrieveIncomingAsyncMessages(),
		IncomingAsyncMessageBytesReceived: node.RetrieveIncomingAsyncMessageBytesReceived(),
	}
}

func newNodeSystemgeInvalidMessageCounters(node *Node.Node) NodeSystemgeInvalidMessageCounters {
	return NodeSystemgeInvalidMessageCounters{
		Name:                                   node.GetName(),
		InvalidMessagesFromIncomingConnections: node.RetrieveInvalidMessagesFromIncomingConnections(),
		InvalidMessagesFromOutgoingConnections: node.RetrieveInvalidMessagesFromOutgoingConnections(),
	}
}

func newNodeSystemgeOutgoingSyncRequestCounters(node *Node.Node) NodeSystemgeOutgoingSyncRequestCounters {
	return NodeSystemgeOutgoingSyncRequestCounters{
		Name:                         node.GetName(),
		OutgoingSyncRequests:         node.RetrieveOutgoingSyncRequests(),
		OutgoingSyncRequestBytesSent: node.RetrieveOutgoingSyncRequestBytesSent(),
	}
}

func newNodeSystemgeOutgoingAsyncMessageCounters(node *Node.Node) NodeSystemgeOutgoingAsyncMessageCounters {
	return NodeSystemgeOutgoingAsyncMessageCounters{
		Name:                          node.GetName(),
		OutgoingAsyncMessages:         node.RetrieveOutgoingAsyncMessages(),
		OutgoingAsyncMessageBytesSent: node.RetrieveOutgoingAsyncMessageBytesSent(),
	}
}

func newNodeSystemgeIncomingConnectionAttemptsCounters(node *Node.Node) NodeSystemgeIncomingConnectionAttemptsCounters {
	return NodeSystemgeIncomingConnectionAttemptsCounters{
		Name:                                   node.GetName(),
		IncomingConnectionAttempts:             node.RetrieveIncomingConnectionAttempts(),
		IncomingConnectionAttemptsSuccessful:   node.RetrieveIncomingConnectionAttemptsSuccessful(),
		IncomingConnectionAttemptsFailed:       node.RetrieveIncomingConnectionAttemptsFailed(),
		IncomingConnectionAttemptBytesSent:     node.RetrieveIncomingConnectionAttemptBytesSent(),
		IncomingConnectionAttemptBytesReceived: node.RetrieveIncomingConnectionAttemptBytesReceived(),
	}
}

func newNodeSystemgeOutgoingConnectionAttemptCounters(node *Node.Node) NodeSystemgeOutgoingConnectionAttemptCounters {
	return NodeSystemgeOutgoingConnectionAttemptCounters{
		Name:                                  node.GetName(),
		OutoingConnectionAttempts:             node.RetrieveOutgoingConnectionAttempts(),
		OutoingConnectionAttemptsSuccessful:   node.RetrieveOutgoingConnectionAttemptsSuccessful(),
		OutoingConnectionAttemptsFailed:       node.RetrieveOutgoingConnectionAttemptsFailed(),
		OutoingConnectionAttemptBytesSent:     node.RetrieveOutgoingConnectionAttemptBytesSent(),
		OutoingConnectionAttemptBytesReceived: node.RetrieveOutgoingConnectionAttemptBytesReceived(),
	}
}

func newNodeSystemgeOutgoingSyncResponsesCounters(node *Node.Node) NodeSystemgeOutgoingSyncResponsesCounters {
	return NodeSystemgeOutgoingSyncResponsesCounters{
		Name:                          node.GetName(),
		OutgoingSyncResponses:         node.RetrieveOutgoingSyncResponses(),
		OutgoingSyncSuccessResponses:  node.RetrieveOutgoingSyncSuccessResponses(),
		OutgoingSyncFailureResponses:  node.RetrieveOutgoingSyncFailureResponses(),
		OutgoingSyncResponseBytesSent: node.RetrieveOutgoingSyncResponseBytesSent(),
	}
}
