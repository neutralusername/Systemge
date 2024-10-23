package accepter

import (
	"errors"

	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

// executes all handlers in order, return error if any handler returns an error
func NewChainedAcceptHandler[T any](handlers ...systemge.AcceptHandlerWithError[T]) systemge.AcceptHandler[T] {
	return func(caller systemge.Connection[T]) {
		for _, handler := range handlers {
			if err := handler(caller); err != nil {
				return
			}
		}
	}
}

type ObtainAcceptHandlerEnqueueConfigs[T any] func(systemge.Connection[T]) (token string, priority uint32, timeoutNs int64)

func NewQueueAcceptHandler[T any](
	priorityTokenQueue *tools.PriorityTokenQueue[systemge.Connection[T]],
	obtainEnqueueConfigs ObtainAcceptHandlerEnqueueConfigs[T],
) systemge.AcceptHandlerWithError[T] {
	return func(caller systemge.Connection[T]) error {
		token, priority, timeoutNs := obtainEnqueueConfigs(caller)
		return priorityTokenQueue.Push(token, caller, priority, timeoutNs)
	}
}

type ObtainIp[T any] func(systemge.Connection[T]) string

func NewControlledAcceptHandler[T any](
	ipRateLimiter *tools.IpRateLimiter,
	blacklist *tools.AccessControlList,
	whitelist *tools.AccessControlList,
	obtainIp ObtainIp[T],
) systemge.AcceptHandlerWithError[T] {
	return func(caller systemge.Connection[T]) error {
		ip := obtainIp(caller)
		if ipRateLimiter != nil && !ipRateLimiter.RegisterConnectionAttempt(ip) {
			return errors.New("rate limited")
		}
		if blacklist != nil && blacklist.Contains(ip) {
			return errors.New("blacklisted")
		}
		if whitelist != nil && whitelist.ElementCount() > 0 && !whitelist.Contains(ip) {
			return errors.New("not whitelisted")
		}
		return nil
	}
}
