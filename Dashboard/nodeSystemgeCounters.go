package Dashboard

import "github.com/neutralusername/Systemge/Node"

type NodeSystemgeCounters struct {
	Name                                   string `json:"name"`
	OutoingConnectionAttempts              uint32 `json:"outgoingConnectionAttempts"`
	OutoingConnectionAttemptsSuccessful    uint32 `json:"outgoingConnectionAttemptsSuccessful"`
	OutoingConnectionAttemptsFailed        uint32 `json:"outgoingConnectionAttemptsFailed"`
	OutoingConnectionAttemptBytesSent      uint64 `json:"outgoingConnectionAttemptBytesSent"`
	OutoingConnectionAttemptBytesReceived  uint64 `json:"outgoingConnectionAttemptBytesReceived"`
	InvalidMessagesFromOutgoingConnections uint32 `json:"invalidMessagesFromOutgoingConnections"`
	IncomingSyncResponses                  uint32 `json:"incomingSyncResponses"`
	IncomingSyncSuccessResponses           uint32 `json:"incomingSyncSuccessResponses"`
	IncomingSyncFailureResponses           uint32 `json:"incomingSyncFailureResponses"`
	IncomingSyncResponseBytesReceived      uint64 `json:"incomingSyncResponseBytesReceived"`
	OutgoingAsyncMessages                  uint32 `json:"outgoingAsyncMessages"`
	OutgoingAsyncMessageBytesSent          uint64 `json:"outgoingAsyncMessageBytesSent"`
	OutgoingSyncRequests                   uint32 `json:"outgoingSyncRequests"`
	OutgoingSyncRequestBytesSent           uint64 `json:"outgoingSyncRequestBytesSent"`
	IncomingConnectionAttempts             uint32 `json:"incomingConnectionAttempts"`
	IncomingConnectionAttemptsSuccessful   uint32 `json:"incomingConnectionAttemptsSuccessful"`
	IncomingConnectionAttemptsFailed       uint32 `json:"incomingConnectionAttemptsFailed"`
	IncomingConnectionAttemptBytesSent     uint64 `json:"incomingConnectionAttemptBytesSent"`
	IncomingConnectionAttemptBytesReceived uint64 `json:"incomingConnectionAttemptBytesReceived"`
	InvalidMessagesFromIncomingConnections uint32 `json:"invalidMessagesFromIncomingConnections"`
	IncomingSyncRequests                   uint32 `json:"incomingSyncRequests"`
	IncomingSyncRequestBytesReceived       uint64 `json:"incomingSyncRequestBytesReceived"`
	IncomingAsyncMessages                  uint32 `json:"incomingAsyncMessages"`
	IncomingAsyncMessageBytesReceived      uint64 `json:"incomingAsyncMessageBytesReceived"`
	OutgoingSyncResponses                  uint32 `json:"outgoingSyncResponses"`
	OutgoingSyncSuccessResponses           uint32 `json:"outgoingSyncSuccessResponses"`
	OutgoingSyncFailureResponses           uint32 `json:"outgoingSyncFailureResponses"`
	OutgoingSyncResponseBytesSent          uint64 `json:"outgoingSyncResponseBytesSent"`
	BytesReceived                          uint64 `json:"bytesReceived"`
	BytesSent                              uint64 `json:"bytesSent"`
}

func newNodeSystemgeCounters(node *Node.Node) NodeSystemgeCounters {
	return NodeSystemgeCounters{
		Name:                                   node.GetName(),
		OutoingConnectionAttempts:              node.RetrieveOutgoingConnectionAttempts(),
		OutoingConnectionAttemptsSuccessful:    node.RetrieveOutgoingConnectionAttemptsSuccessful(),
		OutoingConnectionAttemptsFailed:        node.RetrieveOutgoingConnectionAttemptsFailed(),
		OutoingConnectionAttemptBytesSent:      node.RetrieveOutgoingConnectionAttemptBytesSent(),
		OutoingConnectionAttemptBytesReceived:  node.RetrieveOutgoingConnectionAttemptBytesReceived(),
		InvalidMessagesFromOutgoingConnections: node.RetrieveInvalidMessagesFromOutgoingConnections(),
		IncomingSyncResponses:                  node.RetrieveIncomingSyncResponses(),
		IncomingSyncSuccessResponses:           node.RetrieveIncomingSyncSuccessResponses(),
		IncomingSyncFailureResponses:           node.RetrieveIncomingSyncFailureResponses(),
		IncomingSyncResponseBytesReceived:      node.RetrieveIncomingSyncResponseBytesReceived(),
		OutgoingAsyncMessages:                  node.RetrieveOutgoingAsyncMessages(),
		OutgoingAsyncMessageBytesSent:          node.RetrieveOutgoingAsyncMessageBytesSent(),
		OutgoingSyncRequests:                   node.RetrieveOutgoingSyncRequests(),
		OutgoingSyncRequestBytesSent:           node.RetrieveOutgoingSyncRequestBytesSent(),
		IncomingConnectionAttempts:             node.RetrieveIncomingConnectionAttempts(),
		IncomingConnectionAttemptsSuccessful:   node.RetrieveIncomingConnectionAttemptsSuccessful(),
		IncomingConnectionAttemptsFailed:       node.RetrieveIncomingConnectionAttemptsFailed(),
		IncomingConnectionAttemptBytesSent:     node.RetrieveIncomingConnectionAttemptBytesSent(),
		IncomingConnectionAttemptBytesReceived: node.RetrieveIncomingConnectionAttemptBytesReceived(),
		InvalidMessagesFromIncomingConnections: node.RetrieveInvalidMessagesFromIncomingConnections(),
		IncomingSyncRequests:                   node.RetrieveIncomingSyncRequests(),
		IncomingSyncRequestBytesReceived:       node.RetrieveIncomingSyncRequestBytesReceived(),
		IncomingAsyncMessages:                  node.RetrieveIncomingAsyncMessages(),
		IncomingAsyncMessageBytesReceived:      node.RetrieveIncomingAsyncMessageBytesReceived(),
		OutgoingSyncResponses:                  node.RetrieveOutgoingSyncResponses(),
		OutgoingSyncSuccessResponses:           node.RetrieveOutgoingSyncSuccessResponses(),
		OutgoingSyncFailureResponses:           node.RetrieveOutgoingSyncFailureResponses(),
		OutgoingSyncResponseBytesSent:          node.RetrieveOutgoingSyncResponseBytesSent(),
		BytesReceived:                          node.RetrieveSystemgeBytesReceivedCounter(),
		BytesSent:                              node.RetrieveSystemgeBytesSentCounter(),
	}
}
