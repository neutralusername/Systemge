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
	switch {
	case ipRateLimiter == nil && blockList == nil && accessList == nil:
		return func(connection systemge.Connection[T]) error {
			return nil
		}
	case ipRateLimiter == nil && blockList == nil && accessList != nil:
		return func(connection systemge.Connection[T]) error {
			ip, _, err := net.SplitHostPort(connection.GetAddress())
			if err != nil {
				return err
			}
			if accessList.ElementCount() != 0 && !accessList.Contains(ip) {
				return errors.New("not in access list")
			}
			return nil
		}
	case ipRateLimiter == nil && blockList != nil && accessList == nil:
		return func(connection systemge.Connection[T]) error {
			ip, _, err := net.SplitHostPort(connection.GetAddress())
			if err != nil {
				return err
			}
			if blockList.Contains(ip) {
				return errors.New("blocked")
			}
			return nil
		}
	case ipRateLimiter == nil && blockList != nil && accessList != nil:
		return func(connection systemge.Connection[T]) error {
			ip, _, err := net.SplitHostPort(connection.GetAddress())
			if err != nil {
				return err
			}
			if blockList.Contains(ip) {
				return errors.New("blocked")
			}
			if accessList.ElementCount() != 0 && !accessList.Contains(ip) {
				return errors.New("not in access list")
			}
			return nil
		}

	case ipRateLimiter != nil && blockList == nil && accessList == nil:
		return func(connection systemge.Connection[T]) error {
			ip, _, err := net.SplitHostPort(connection.GetAddress())
			if err != nil {
				return err
			}
			if !ipRateLimiter.RegisterConnectionAttempt(ip) {
				return errors.New("rate limited")
			}
			return nil
		}
	case ipRateLimiter != nil && blockList == nil && accessList != nil:
		return func(connection systemge.Connection[T]) error {
			ip, _, err := net.SplitHostPort(connection.GetAddress())
			if err != nil {
				return err
			}
			if !ipRateLimiter.RegisterConnectionAttempt(ip) {
				return errors.New("rate limited")
			}
			if accessList.ElementCount() != 0 && !accessList.Contains(ip) {
				return errors.New("rate limited or not in access list")
			}
			return nil
		}
	case ipRateLimiter != nil && blockList != nil && accessList == nil:
		return func(connection systemge.Connection[T]) error {
			ip, _, err := net.SplitHostPort(connection.GetAddress())
			if err != nil {
				return err
			}
			if !ipRateLimiter.RegisterConnectionAttempt(ip) {
				return errors.New("rate limited")
			}
			if blockList.Contains(ip) {
				return errors.New("rate limited or blocked")
			}
			return nil
		}
	case ipRateLimiter != nil && blockList != nil && accessList != nil:
		return func(connection systemge.Connection[T]) error {
			ip, _, err := net.SplitHostPort(connection.GetAddress())
			if err != nil {
				return err
			}
			if !ipRateLimiter.RegisterConnectionAttempt(ip) {
				return errors.New("rate limited")
			}
			if blockList.Contains(ip) {
				return errors.New("rate limited or blocked")
			}
			if accessList.ElementCount() != 0 && !accessList.Contains(ip) {
				return nil
			}
			return errors.New("rate limited or blocked or not in access list")
		}
	default:
		return nil
	}
}

/* func NewPasswordHandler[T any](
	getCurrentPassword func(connection systemge.Connection[T]) string,
	unmarshalPassword func(password T) (string, error),
	timeoutNs int64,
) systemge.AcceptHandlerWithError[T] {
	return func(connection systemge.Connection[T]) error {
		currentPassword := getCurrentPassword(connection)
		if currentPassword == "" {
			return errors.New("no password set")
		}
		password, err := connection.Read(timeoutNs)
		if err != nil {
			return err
		}
		unmarshaledPassword, err := unmarshalPassword(password)
		if err != nil {
			return err
		}
		if unmarshaledPassword != currentPassword {
			return errors.New("wrong password")
		}
		return nil
	}
} */

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
