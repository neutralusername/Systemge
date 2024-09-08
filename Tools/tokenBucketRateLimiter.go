package Tools

import (
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
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

func (rateLimiter *TokenBucketRateLimiter) Close() {
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

func (rateLimiter *TokenBucketRateLimiter) GetRefillInterval() uint64 {
	rateLimiter.mutex.Lock()
	defer rateLimiter.mutex.Unlock()
	return rateLimiter.refillIntervalMs
}
func (rateLimiter *TokenBucketRateLimiter) SetRefillInterval(interval uint64) {
	rateLimiter.mutex.Lock()
	defer rateLimiter.mutex.Unlock()
	rateLimiter.refillIntervalMs = interval
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

func (rateLimiter *TokenBucketRateLimiter) GetCommands() Commands.Handlers {
	commands := Commands.Handlers{}
	commands["consume"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("consume expects 1 argument", nil)
		}
		if !rateLimiter.Consume(Helpers.StringToUint64(args[0])) {
			return "", Error.New("failed to consume", nil)
		}
		return "success", nil
	}
	commands["refill"] = func(args []string) (string, error) {
		rateLimiter.Refill()
		return "success", nil
	}
	commands["getRefillRate"] = func(args []string) (string, error) {
		return Helpers.Uint64ToString(rateLimiter.GetRefillRate()), nil
	}
	commands["setRefillRate"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("setRefillRate expects 1 argument", nil)
		}
		rateLimiter.SetRefillRate(Helpers.StringToUint64(args[0]))
		return "success", nil
	}
	commands["getRefillInterval"] = func(args []string) (string, error) {
		return Helpers.Uint64ToString(rateLimiter.GetRefillInterval()), nil
	}
	commands["setRefillInterval"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("setRefillInterval expects 1 argument", nil)
		}
		rateLimiter.SetRefillInterval(Helpers.StringToUint64(args[0]))
		return "success", nil
	}
	commands["getMaxBucketSize"] = func(args []string) (string, error) {
		return Helpers.Uint64ToString(rateLimiter.GetMaxBucketSize()), nil
	}
	commands["setMaxBucketSize"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("setMaxBucketSize expects 1 argument", nil)
		}
		rateLimiter.SetMaxBucketSize(Helpers.StringToUint64(args[0]))
		return "success", nil
	}
	return commands
}
