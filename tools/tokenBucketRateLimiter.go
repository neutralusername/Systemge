package tools

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/helpers"
)

type TokenBucketRateLimiter struct {
	bucket           uint64
	maxBucketSize    uint64
	refillRate       uint64
	refillIntervalNs int64
	active           bool
	mutex            sync.Mutex
}

func NewTokenBucketRateLimiter(config *Config.TokenBucketRateLimiter) *TokenBucketRateLimiter {
	rateLimiter := &TokenBucketRateLimiter{
		bucket:           config.InitialBucketSize,
		maxBucketSize:    config.MaxBucketSize,
		refillRate:       config.RefillRate,
		refillIntervalNs: config.RefillIntervalNs,
		active:           true,
	}
	if config.RefillIntervalNs > 0 && config.RefillRate > 0 {
		go rateLimiter.refillRoutine()
	}
	return rateLimiter
}

func (rateLimiter *TokenBucketRateLimiter) Close() {
	rateLimiter.active = false
}

func (rateLimiter *TokenBucketRateLimiter) refillRoutine() {
	time.Sleep(time.Duration(rateLimiter.refillIntervalNs) * time.Nanosecond)
	for rateLimiter.active {
		rateLimiter.mutex.Lock()
		rateLimiter.bucket += rateLimiter.refillRate
		if rateLimiter.bucket > rateLimiter.maxBucketSize {
			rateLimiter.bucket = rateLimiter.maxBucketSize
		}
		rateLimiter.mutex.Unlock()
		time.Sleep(time.Duration(rateLimiter.refillIntervalNs) * time.Nanosecond)
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

func (rateLimiter *TokenBucketRateLimiter) Refill() {
	rateLimiter.mutex.Lock()
	defer rateLimiter.mutex.Unlock()
	rateLimiter.bucket = rateLimiter.maxBucketSize
}

func (rateLimiter *TokenBucketRateLimiter) GetRefillRate() uint64 {
	rateLimiter.mutex.Lock()
	defer rateLimiter.mutex.Unlock()
	return rateLimiter.refillRate
}
func (rateLimiter *TokenBucketRateLimiter) SetRefillRate(rate uint64) {
	rateLimiter.mutex.Lock()
	defer rateLimiter.mutex.Unlock()
	rateLimiter.refillRate = rate
}

func (rateLimiter *TokenBucketRateLimiter) GetRefillIntervalNs() int64 {
	rateLimiter.mutex.Lock()
	defer rateLimiter.mutex.Unlock()
	return rateLimiter.refillIntervalNs
}
func (rateLimiter *TokenBucketRateLimiter) SetRefillIntervalNs(intervalNs int64) {
	rateLimiter.mutex.Lock()
	defer rateLimiter.mutex.Unlock()
	rateLimiter.refillIntervalNs = intervalNs
}

func (rateLimiter *TokenBucketRateLimiter) GetMaxBucketSize() uint64 {
	rateLimiter.mutex.Lock()
	defer rateLimiter.mutex.Unlock()
	return rateLimiter.maxBucketSize
}
func (rateLimiter *TokenBucketRateLimiter) SetMaxBucketSize(size uint64) {
	rateLimiter.mutex.Lock()
	defer rateLimiter.mutex.Unlock()
	rateLimiter.maxBucketSize = size
}

func (rateLimiter *TokenBucketRateLimiter) GetDefaultCommands() Handlers {
	commands := Handlers{}
	commands["consume"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", errors.New("consume expects 1 argument")
		}
		if !rateLimiter.Consume(helpers.StringToUint64(args[0])) {
			return "", errors.New("failed to consume")
		}
		return "success", nil
	}
	commands["refill"] = func(args []string) (string, error) {
		rateLimiter.Refill()
		return "success", nil
	}
	commands["getRefillRate"] = func(args []string) (string, error) {
		return helpers.Uint64ToString(rateLimiter.GetRefillRate()), nil
	}
	commands["setRefillRate"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", errors.New("setRefillRate expects 1 argument")
		}
		rateLimiter.SetRefillRate(helpers.StringToUint64(args[0]))
		return "success", nil
	}
	commands["getRefillInterval"] = func(args []string) (string, error) {
		return helpers.Int64ToString(rateLimiter.GetRefillIntervalNs()), nil
	}
	commands["setRefillInterval"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", errors.New("setRefillInterval expects 1 argument")
		}
		rateLimiter.SetRefillIntervalNs(helpers.StringToInt64(args[0]))
		return "success", nil
	}
	commands["getMaxBucketSize"] = func(args []string) (string, error) {
		return helpers.Uint64ToString(rateLimiter.GetMaxBucketSize()), nil
	}
	commands["setMaxBucketSize"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", errors.New("setMaxBucketSize expects 1 argument")
		}
		rateLimiter.SetMaxBucketSize(helpers.StringToUint64(args[0]))
		return "success", nil
	}
	return commands
}
