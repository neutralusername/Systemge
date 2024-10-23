package accepter

import (
	"errors"
	"net"

	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func NewAccessControllHandler[T any](
	ipRateLimiter *tools.IpRateLimiter,
	blockList *tools.AccessControlList,
	accessList *tools.AccessControlList,
) systemge.AcceptHandlerWithError[T] {
	return func(connection systemge.Connection[T]) error {
		ip, _, err := net.SplitHostPort(connection.GetAddress())
		if err != nil {
			return err
		}
		if ipRateLimiter != nil && !ipRateLimiter.RegisterConnectionAttempt(ip) {
			return errors.New("rate limited")
		}
		if blockList != nil && blockList.Contains(ip) {
			return errors.New("blocked")
		}
		if accessList != nil && accessList.ElementCount() > 0 && !accessList.Contains(ip) {
			return errors.New("not in access list")
		}
		return nil
	}
}

/*
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
*/
