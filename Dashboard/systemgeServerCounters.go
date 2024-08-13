package Dashboard

import "github.com/neutralusername/Systemge/Node"

type SystemgeServerCounters struct {
	Name                    string `json:"name"`
	BytesReceived           uint64 `json:"bytesReceived"`
	BytesSent               uint64 `json:"bytesSent"`
	InvalidMessagesReceived uint32 `json:"invalidMessagesReceived"`
}

func newNodeSystemgeServerCounters(node *Node.Node) SystemgeServerCounters {
	return SystemgeServerCounters{
		Name:                    node.GetName(),
		BytesReceived:           node.RetrieveSystemgeServerBytesReceived(),
		BytesSent:               node.RetrieveSystemgeServerBytesSent(),
		InvalidMessagesReceived: node.RetrieveSystemgeServerInvalidMessagesReceived(),
	}
}

type SystemgeServerRateLimitCounters struct {
	Name                       string `json:"name"`
	MessageRateLimiterExceeded uint32 `json:"messageRateLimiterExceeded"`
	ByteRateLimiterExceeded    uint32 `json:"byteRateLimiterExceeded"`
}

func newNodeSystemgeServerRateLimitCounters(node *Node.Node) SystemgeServerRateLimitCounters {
	return SystemgeServerRateLimitCounters{
		Name:                       node.GetName(),
		MessageRateLimiterExceeded: node.RetrieveSystemgeServerMessageRateLimiterExceeded(),
		ByteRateLimiterExceeded:    node.RetrieveSystemgeServerByteRateLimiterExceeded(),
	}
}

type SystemgeServerConnectionCounters struct {
	Name                           string `json:"name"`
	ConnectionAttempts             uint32 `json:"connectionAttempts"`
	ConnectionAttemptsSuccessful   uint32 `json:"connectionAttemptsSuccessful"`
	ConnectionAttemptsFailed       uint32 `json:"connectionAttemptsFailed"`
	ConnectionAttemptBytesSent     uint64 `json:"connectionAttemptBytesSent"`
	ConnectionAttemptBytesReceived uint64 `json:"connectionAttemptBytesReceived"`
}

func newNodeSystemgeServerConnectionCounters(node *Node.Node) SystemgeServerConnectionCounters {
	return SystemgeServerConnectionCounters{
		Name:                           node.GetName(),
		ConnectionAttempts:             node.RetrieveSystemgeServerConnectionAttempts(),
		ConnectionAttemptsSuccessful:   node.RetrieveSystemgeServerConnectionAttemptsSuccessful(),
		ConnectionAttemptsFailed:       node.RetrieveSystemgeServerConnectionAttemptsFailed(),
		ConnectionAttemptBytesSent:     node.RetrieveSystemgeServerConnectionAttemptBytesSent(),
		ConnectionAttemptBytesReceived: node.RetrieveSystemgeServerConnectionAttemptBytesReceived(),
	}
}

type SystemgeServerSyncResponseCounters struct {
	Name                     string `json:"name"`
	SyncSuccessResponsesSent uint32 `json:"syncSuccessResponsesSent"`
	SyncFailureResponsesSent uint32 `json:"syncFailureResponsesSent"`
	SyncResponseBytesSent    uint64 `json:"syncResponseBytesSent"`
}

func newNodeSystemgeServerSyncResponseCounters(node *Node.Node) SystemgeServerSyncResponseCounters {
	return SystemgeServerSyncResponseCounters{
		Name:                     node.GetName(),
		SyncSuccessResponsesSent: node.RetrieveSystemgeServerSyncSuccessResponsesSent(),
		SyncFailureResponsesSent: node.RetrieveSystemgeServerSyncFailureResponsesSent(),
		SyncResponseBytesSent:    node.RetrieveSystemgeServerSyncResponseBytesSent(),
	}
}

type SystemgeServerAsyncMessageCounters struct {
	Name                      string `json:"name"`
	AsyncMessagesReceived     uint32 `json:"asyncMessagesReceived"`
	AsyncMessageBytesReceived uint64 `json:"asyncMessageBytesReceived"`
}

func newNodeSystemgeServerAsyncMessageCounters(node *Node.Node) SystemgeServerAsyncMessageCounters {
	return SystemgeServerAsyncMessageCounters{
		Name:                      node.GetName(),
		AsyncMessagesReceived:     node.RetrieveSystemgeServerAsyncMessagesReceived(),
		AsyncMessageBytesReceived: node.RetrieveSystemgeServerAsyncMessageBytesReceived(),
	}
}

type SystemgeServerSyncRequestCounters struct {
	Name                     string `json:"name"`
	SyncRequestsReceived     uint32 `json:"syncRequestsReceived"`
	SyncRequestBytesReceived uint64 `json:"syncRequestBytesReceived"`
}

func newNodeSystemgeServerSyncRequestCounters(node *Node.Node) SystemgeServerSyncRequestCounters {
	return SystemgeServerSyncRequestCounters{
		Name:                     node.GetName(),
		SyncRequestsReceived:     node.RetrieveSystemgeServerSyncRequestsReceived(),
		SyncRequestBytesReceived: node.RetrieveSystemgeServerSyncRequestBytesReceived(),
	}
}

type SystemgeServerTopicCounters struct {
	Name            string `json:"name"`
	TopicAddSent    uint32 `json:"topicAddSent"`
	TopicRemoveSent uint32 `json:"topicRemoveSent"`
}

func newNodeSystemgeServerTopicCounters(node *Node.Node) SystemgeServerTopicCounters {
	return SystemgeServerTopicCounters{
		Name:            node.GetName(),
		TopicAddSent:    node.RetrieveSystemgeServerTopicAddSent(),
		TopicRemoveSent: node.RetrieveSystemgeServerTopicRemoveSent(),
	}
}
