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
