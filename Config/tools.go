package Config

import (
	"encoding/json"
)

type Mailer struct {
	SmtpHost       string   `json:"smtpHost"`       // *required*
	SmtpPort       uint16   `json:"smtpPort"`       // *required*
	SenderEmail    string   `json:"senderEmail"`    // *required*
	SenderPassword string   `json:"senderPassword"` // *required*
	Recipients     []string `json:"recipients"`
}

func UnmarshalMailer(data string) *Mailer {
	var mailer Mailer
	err := json.Unmarshal([]byte(data), &mailer)
	if err != nil {
		return nil
	}
	return &mailer
}

type TokenBucketRateLimiter struct {
	InitialBucketSize uint64 `json:"initialBucketSize"` // default: 0
	MaxBucketSize     uint64 `json:"maxBucketSize"`     // default: 0
	RefillRate        uint64 `json:"refillRate"`        // default: 0 == no refill
	RefillIntervalMs  uint64 `json:"refillIntervalMs"`  // default: 0 == no refill
}

func UnmarshalRateLimiter(data string) *TokenBucketRateLimiter {
	var rateLimiterConfig TokenBucketRateLimiter
	err := json.Unmarshal([]byte(data), &rateLimiterConfig)
	if err != nil {
		return nil
	}
	return &rateLimiterConfig
}

type IpRateLimiter struct {
	MaxAttempts         uint32 `json:"maxAttempts"`         // default: 1
	AttemptTimeWindowMs uint32 `json:"attemptTimeWindowMs"` // default: 1
	CleanupIntervalMs   uint32 `json:"cleanupIntervalMs"`   // default: 1000ms
}

func UnmarshalIpRateLimiter(data string) *IpRateLimiter {
	var ipRateLimiterConfig IpRateLimiter
	err := json.Unmarshal([]byte(data), &ipRateLimiterConfig)
	if err != nil {
		return nil
	}
	return &ipRateLimiterConfig
}

type SessionManager struct {
	SessionLifetimeMs      uint64 `json:"sessionLifetimeMs"`      // default: 0 == no expiration
	SessionIdLength        uint32 `json:"sessionIdLength"`        // default: 32
	MaxSessionsPerIdentity uint32 `json:"maxSessionsPerIdentity"` // default: 0 == no limit
	MaxIdentities          uint32 `json:"maxIdentities"`          // default: 0 == no limit
	MaxIdentityLength      uint32 `json:"maxIdentityLength"`      // default: 0 == no limit
	MinIdentityLength      uint32 `json:"minIdentityLength"`      // default: 0 == no limit
	SessionIdAlphabet      string `json:"sessionIdAlphabet"`      // default: alphanumeric
}

func UnmarshalSessionManager(data string) *SessionManager {
	var sessionManagerConfig SessionManager
	err := json.Unmarshal([]byte(data), &sessionManagerConfig)
	if err != nil {
		return nil
	}
	return &sessionManagerConfig
}

type TopicManager struct {
	TopicQueueSize     uint32 `json:"topicQueueSize"`     // default: 0
	QueueSize          uint32 `json:"queueSize"`          // default: 0
	ConcurrentCalls    bool   `json:"concurrentCalls"`    // default: false
	QueueBlocking      bool   `json:"queueBlocking"`      // default: false // if false, will drop calls if queue is full. will wait if true
	TopicQueueBlocking bool   `json:"topicQueueBlocking"` // default: false // if false, will drop calls if topicQueue is full. will wait if true
	DeadlineMs         uint64 `json:"deadlineMs"`         // default: 0 == no deadline
}

func UnmarshalTopicManager(data string) *TopicManager {
	var topicManagerConfig TopicManager
	err := json.Unmarshal([]byte(data), &topicManagerConfig)
	if err != nil {
		return nil
	}
	return &topicManagerConfig
}

type SyncManager struct {
	MaxTokenLength    int `json:"maxTokenLength"`    // default: 0 == no limit
	MinTokenLength    int `json:"minTokenLength"`    // default: 0 == no limit
	MaxActiveRequests int `json:"maxActiveRequests"` // default: 0 == no limit
}
