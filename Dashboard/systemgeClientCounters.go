package Dashboard

import "github.com/neutralusername/Systemge/Node"

type SystemgeClientCounters struct {
	Name                    string `json:"name"`
	BytesReceived           uint64 `json:"bytesReceived"`
	BytesSent               uint64 `json:"bytesSent"`
	InvalidMessagesReceived uint32 `json:"invalidMessagesReceived"`
}

func newNodeSystemgeClientCounters(node *Node.Node) SystemgeClientCounters {
	return SystemgeClientCounters{
		Name:                    node.GetName(),
		BytesReceived:           node.RetrieveSystemgeClientBytesReceived(),
		BytesSent:               node.RetrieveSystemgeClientBytesSent(),
		InvalidMessagesReceived: node.RetrieveSystemgeClientInvalidMessagesReceived(),
	}
}

type SystemgeClientRateLimitCounters struct {
	Name                       string `json:"name"`
	MessageRateLimiterExceeded uint32 `json:"messageRateLimiterExceeded"`
	ByteRateLimiterExceeded    uint32 `json:"byteRateLimiterExceeded"`
}

func newNodeSystemgeClientRateLimitCounters(node *Node.Node) SystemgeClientRateLimitCounters {
	return SystemgeClientRateLimitCounters{
		Name:                       node.GetName(),
		MessageRateLimiterExceeded: node.RetrieveSystemgeClientMessageRateLimiterExceeded(),
		ByteRateLimiterExceeded:    node.RetrieveSystemgeClientByteRateLimiterExceeded(),
	}
}

type SystemgeClientConnectionCounters struct {
	Name                           string `json:"name"`
	ConnectionAttempts             uint32 `json:"connectionAttempts"`
	ConnectionAttemptsSuccessful   uint32 `json:"connectionAttemptsSuccessful"`
	ConnectionAttemptsFailed       uint32 `json:"connectionAttemptsFailed"`
	ConnectionAttemptBytesSent     uint64 `json:"connectionAttemptBytesSent"`
	ConnectionAttemptBytesReceived uint64 `json:"connectionAttemptBytesReceived"`
}

func newNodeSystemgeClientConnectionCounters(node *Node.Node) SystemgeClientConnectionCounters {
	return SystemgeClientConnectionCounters{
		Name:                           node.GetName(),
		ConnectionAttempts:             node.RetrieveSystemgeClientConnectionAttempts(),
		ConnectionAttemptsSuccessful:   node.RetrieveSystemgeClientConnectionAttemptsSuccessful(),
		ConnectionAttemptsFailed:       node.RetrieveSystemgeClientConnectionAttemptsFailed(),
		ConnectionAttemptBytesSent:     node.RetrieveSystemgeClientConnectionAttemptBytesSent(),
		ConnectionAttemptBytesReceived: node.RetrieveSystemgeClientConnectionAttemptBytesReceived(),
	}
}

type SystemgeClientSyncResponseCounters struct {
	Name                         string `json:"name"`
	SyncSuccessResponsesReceived uint32 `json:"syncSuccessResponsesReceived"`
	SyncFailureResponsesReceived uint32 `json:"syncFailureResponsesReceived"`
	SyncResponseBytesReceived    uint64 `json:"syncResponseBytesReceived"`
}

func newNodeSystemgeClientSyncResponseCounters(node *Node.Node) SystemgeClientSyncResponseCounters {
	return SystemgeClientSyncResponseCounters{
		Name:                         node.GetName(),
		SyncSuccessResponsesReceived: node.RetrieveSystemgeClientSyncSuccessResponsesReceived(),
		SyncFailureResponsesReceived: node.RetrieveSystemgeClientSyncFailureResponsesReceived(),
		SyncResponseBytesReceived:    node.RetrieveSystemgeClientSyncResponseBytesReceived(),
	}
}

type SystemgeClientAsyncMessageCounters struct {
	Name                  string `json:"name"`
	AsyncMessagesSent     uint32 `json:"asyncMessagesSent"`
	AsyncMessageBytesSent uint64 `json:"asyncMessageBytesSent"`
}

func newNodeSystemgeClientAsyncMessageCounters(node *Node.Node) SystemgeClientAsyncMessageCounters {
	return SystemgeClientAsyncMessageCounters{
		Name:                  node.GetName(),
		AsyncMessagesSent:     node.RetrieveSystemgeClientAsyncMessagesSent(),
		AsyncMessageBytesSent: node.RetrieveSystemgeClientAsyncMessageBytesSent(),
	}
}

type SystemgeClientSyncRequestCounters struct {
	Name                 string `json:"name"`
	SyncRequestsSent     uint32 `json:"syncRequestsSent"`
	SyncRequestBytesSent uint64 `json:"syncRequestBytesSent"`
}

func newNodeSystemgeClientSyncRequestCounters(node *Node.Node) SystemgeClientSyncRequestCounters {
	return SystemgeClientSyncRequestCounters{
		Name:                 node.GetName(),
		SyncRequestsSent:     node.RetrieveSystemgeClientSyncRequestsSent(),
		SyncRequestBytesSent: node.RetrieveSystemgeClientSyncRequestBytesSent(),
	}
}

type SystemgeClientTopicCounters struct {
	Name                string `json:"name"`
	TopicAddReceived    uint32 `json:"topicAddReceived"`
	TopicRemoveReceived uint32 `json:"topicRemoveReceived"`
}

func newNodeSystemgeClientTopicCounters(node *Node.Node) SystemgeClientTopicCounters {
	return SystemgeClientTopicCounters{
		Name:                node.GetName(),
		TopicAddReceived:    node.RetrieveSystemgeClientTopicAddReceived(),
		TopicRemoveReceived: node.RetrieveSystemgeClientTopicRemoveReceived(),
	}
}
