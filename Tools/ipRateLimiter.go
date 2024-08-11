package Tools

import (
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Config"
)

type IpRateLimiter struct {
	connections     map[string][]time.Time
	mutex           sync.Mutex
	active          bool
	timeWindow      time.Duration
	maxAttempts     int
	cleanupInterval time.Duration
}

func (rl *IpRateLimiter) Stop() {
	rl.active = false
}

func NewIpRateLimiter(config *Config.IpRateLimiter) *IpRateLimiter {
	if config.MaxAttempts < 1 {
		config.MaxAttempts = 1
	}
	if config.AttemptTimeWindowMs < 1 {
		config.AttemptTimeWindowMs = 1
	}
	if config.CleanupIntervalMs < 1 {
		config.CleanupIntervalMs = 1
	}
	rl := &IpRateLimiter{
		connections:     make(map[string][]time.Time),
		active:          true,
		timeWindow:      time.Duration(config.AttemptTimeWindowMs) * time.Millisecond,
		maxAttempts:     config.MaxAttempts,
		cleanupInterval: time.Duration(config.CleanupIntervalMs) * time.Millisecond,
	}
	go rl.cleanupOldEntries()
	return rl
}

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
