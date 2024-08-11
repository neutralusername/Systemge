package Tools

import (
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Config"
)

type TokenBucketRateLimiter struct {
	bucket           uint64
	maxBucketSize    uint64
	refillRate       uint64
	refillIntervalMs uint64
	active           bool
	mutex            sync.Mutex
}

func NewTokenBucketRateLimiter(config *Config.TokenBucketRateLimiter) *TokenBucketRateLimiter {
	rateLimiter := &TokenBucketRateLimiter{
		bucket:           config.InitialBucketSize,
		maxBucketSize:    config.MaxBucketSize,
		refillRate:       config.RefillRate,
		refillIntervalMs: config.RefillIntervalMs,
		active:           true,
	}
	if config.RefillIntervalMs > 0 && config.RefillRate > 0 {
		go rateLimiter.refillRoutine()
	}
	return rateLimiter
}

func (rateLimiter *TokenBucketRateLimiter) Stop() {
	rateLimiter.active = false
}

func (rateLimiter *TokenBucketRateLimiter) refillRoutine() {
	time.Sleep(time.Duration(rateLimiter.refillIntervalMs) * time.Millisecond)
	for rateLimiter.active {
		rateLimiter.mutex.Lock()
		rateLimiter.bucket += rateLimiter.refillRate
		if rateLimiter.bucket > rateLimiter.maxBucketSize {
			rateLimiter.bucket = rateLimiter.maxBucketSize
		}
		rateLimiter.mutex.Unlock()
		time.Sleep(time.Duration(rateLimiter.refillIntervalMs) * time.Millisecond)
	}
}

func (rateLimiter *TokenBucketRateLimiter) Consume(amount uint64) bool {
	rateLimiter.mutex.Lock()
	defer rateLimiter.mutex.Unlock()
	if rateLimiter.bucket < amount {
		return false
	}
	rateLimiter.bucket -= amount
	return true
}
