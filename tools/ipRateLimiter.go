package tools

import (
	"sync"
	"time"

	"github.com/neutralusername/systemge/configs"
)

type IpRateLimiter struct {
	connections     map[string][]time.Time
	mutex           sync.Mutex
	active          bool
	timeWindow      time.Duration
	maxAttempts     int
	cleanupInterval time.Duration
}

func (rl *IpRateLimiter) Close() {
	rl.active = false
}

func NewIpRateLimiter(config *configs.IpRateLimiter) *IpRateLimiter {
	if config.MaxAttempts < 1 {
		config.MaxAttempts = 1
	}
	if config.AttemptTimeWindowNs < 1 {
		config.AttemptTimeWindowNs = 1
	}
	if config.CleanupIntervalNs < 1 {
		config.CleanupIntervalNs = 1000
	}
	rl := &IpRateLimiter{
		connections:     make(map[string][]time.Time),
		active:          true,
		timeWindow:      time.Duration(config.AttemptTimeWindowNs) * time.Nanosecond,
		maxAttempts:     config.MaxAttempts,
		cleanupInterval: time.Duration(config.CleanupIntervalNs) * time.Nanosecond,
	}
	go rl.cleanupOldEntries()
	return rl
}

// Returns true if the connection attempt is allowed, false otherwise
func (rl *IpRateLimiter) RegisterConnectionAttempt(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	now := time.Now()
	attempts, exists := rl.connections[ip]
	if !exists {
		rl.connections[ip] = []time.Time{now}
		return true
	}
	recentAttempts := []time.Time{}
	for _, t := range attempts {
		if now.Sub(t) <= rl.timeWindow {
			recentAttempts = append(recentAttempts, t)
		}
	}
	rl.connections[ip] = append(recentAttempts, now)
	return len(recentAttempts) < rl.maxAttempts
}

func (rl *IpRateLimiter) cleanupOldEntries() {
	for rl.active {
		time.Sleep(rl.cleanupInterval)
		rl.mutex.Lock()
		now := time.Now()
		for ip, attempts := range rl.connections {
			recentAttempts := []time.Time{}
			for _, t := range attempts {
				if now.Sub(t) <= rl.timeWindow {
					recentAttempts = append(recentAttempts, t)
				}
			}
			if len(recentAttempts) == 0 {
				delete(rl.connections, ip)
			} else {
				rl.connections[ip] = recentAttempts
			}
		}
		rl.mutex.Unlock()
	}
}
