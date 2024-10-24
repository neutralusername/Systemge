package accepter

import (
	"errors"
	"net"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

// executes all handlers in order, return error if any handler returns an error
func NewChainedHandler[T any](handlers ...systemge.AcceptHandlerWithError[T]) systemge.AcceptHandlerWithError[T] {
	return func(caller systemge.Connection[T]) error {
		for _, handler := range handlers {
			if err := handler(caller); err != nil {
				return err
			}
		}
		return nil
	}
}

// rejects incoming connections based on ipRateLimiter, blockList and accessList.
// arguments may be nil if not needed.
func NewAccessControlHandler[T any](
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

// executes getCurrentPassword with connection as argument, reads message from connection, unmarshals it and compares with current password.
// optionally writes requestMessage to connection before attempting to read password.
// returns nil error if passwords match.
// note: be wary of coordination between reads and writes on both ends.
func NewAuthenticationHandler[T any](
	getCurrentPassword func(connection systemge.Connection[T]) string,
	unmarshalPassword func(password T) (string, error),
	requestMessage func(connection systemge.Connection[T]) (T, bool),
	writeTimeoutNs int64,
	readTimeoutNs int64,
) systemge.AcceptHandlerWithError[T] {
	return func(connection systemge.Connection[T]) error {
		currentPassword := getCurrentPassword(connection)
		if currentPassword == "" {
			return nil
		}
		if requestMessage, ok := requestMessage(connection); ok {
			if err := connection.Write(requestMessage, writeTimeoutNs); err != nil {
				return err
			}
		}
		data, err := connection.Read(readTimeoutNs)
		if err != nil {
			return err
		}
		password, err := unmarshalPassword(data)
		if err != nil {
			return err
		}
		if password != currentPassword {
			return errors.New("wrong password")
		}
		return nil
	}
}

// reads data from connection, executes readHandler and writes result back to connection and closes connection afterwards.
func NewSingleReadAsyncHandler[T any](
	readerConfig *configs.ReaderAsync,
	readHandler systemge.ReadHandler[T],
	stopChannel <-chan struct{},
) systemge.AcceptHandlerWithError[T] {
	return func(connection systemge.Connection[T]) error {
		select {
		case <-stopChannel:
			connection.SetReadDeadline(1)
			// routine was stopped
			return errors.New("routine was stopped")

		case <-connection.GetCloseChannel():
			// ending routine due to connection close
			return errors.New("connection was closed")

		case data, ok := <-helpers.ChannelCall(func() (T, error) { return connection.Read(readerConfig.ReadTimeoutNs) }):
			if !ok {
				// do smthg with the error
				return errors.New("error reading data")
			}
			readHandler(data, connection)
			connection.Close()
			return nil
		}
	}
}

// reads data from connection, executes readHandler and writes result back to connection and closes connection afterwards.
func NewSingleReadSyncHandler[T any](
	readerConfig *configs.ReaderSync,
	readHandler systemge.ReadHandlerWithResult[T],
	stopChannel <-chan struct{},
) systemge.AcceptHandlerWithError[T] {
	return func(connection systemge.Connection[T]) error {
		select {
		case <-stopChannel:
			connection.SetReadDeadline(1)
			// routine was stopped
			return errors.New("routine was stopped")

		case <-connection.GetCloseChannel():
			// ending routine due to connection close
			return errors.New("connection was closed")

		case data, ok := <-helpers.ChannelCall(func() (T, error) { return connection.Read(readerConfig.ReadTimeoutNs) }):
			if !ok {
				// do smthg with the error
				return errors.New("error reading data")
			}
			result, err := readHandler(data, connection)
			connection.Close()
			if err != nil {
				return err
			}
			connection.Write(result, readerConfig.WriteTimeoutNs)

			return nil
		}
	}
}

// adds connection to connection manager and removes it when connection is closed.
func AcceptConnectionManagerHandler[T any](
	connectionManager *systemge.ConnectionManager[T],
	removeOnClose bool,
) systemge.AcceptHandlerWithError[T] {
	if removeOnClose {
		return func(connection systemge.Connection[T]) error {
			_, err := connectionManager.Add(connection)
			go func() {
				<-connection.GetCloseChannel()
				connectionManager.Remove(connection)
			}()
			return err
		}
	} else {
		return func(connection systemge.Connection[T]) error {
			_, err := connectionManager.Add(connection)
			return err
		}
	}
}

// adds connection to connection manager and removes it when connection is closed.
// could be used multiple times with different managers and getId functions (userId, groupId, etc.).
func AcceptConnectionManagerIdHandler[T any](
	connectionManager *systemge.ConnectionManager[T],
	removeOnClose bool,
	getId func(connection systemge.Connection[T]) string,
) systemge.AcceptHandlerWithError[T] {
	if removeOnClose {
		return func(connection systemge.Connection[T]) error {
			if err := connectionManager.AddId(getId(connection), connection); err != nil {
				return err
			}
			go func() {
				<-connection.GetCloseChannel()
				connectionManager.Remove(connection)
			}()
			return nil
		}
	} else {
		return func(connection systemge.Connection[T]) error {
			return connectionManager.AddId(getId(connection), connection)
		}
	}
}

// executes provided function once connection closes.
func OnCloseHandler[T any](
	onClose func(connection systemge.Connection[T]),
) systemge.AcceptHandlerWithError[T] {
	return func(connection systemge.Connection[T]) error {
		go func() {
			<-connection.GetCloseChannel()
			onClose(connection)
		}()
		return nil
	}
}

type ObtainEnqueueConfigs[T any] func(systemge.Connection[T]) (token string, priority uint32, timeoutNs int64)

// enqueues incoming connections.
func NewQueueHandler[T any](
	priorityTokenQueue *tools.PriorityTokenQueue[systemge.Connection[T]],
	obtainEnqueueConfigs ObtainEnqueueConfigs[T],
) systemge.AcceptHandlerWithError[T] {

	return func(connection systemge.Connection[T]) error {
		token, priority, timeoutNs := obtainEnqueueConfigs(connection)
		return priorityTokenQueue.Push(token, connection, priority, timeoutNs)
	}
}

// repeatdedly dequeues connections and executes provided handler.
func NewDequeueRoutine[T any](
	priorityTokenQueue *tools.PriorityTokenQueue[systemge.Connection[T]],
	acceptHandler systemge.AcceptHandlerWithError[T],
	dequeueRoutineConfig *configs.Routine,
) (*tools.Routine, error) {
	routine, err := tools.NewRoutine(
		func(stopChannel <-chan struct{}) {
			for {
				select {
				case <-stopChannel:
					return
				case connection, ok := <-priorityTokenQueue.PopChannel():
					if !ok {
						return
					}
					err := acceptHandler(connection)
					if err != nil {

					}
				}
			}
		},
		dequeueRoutineConfig,
	)
	if err != nil {
		return nil, err
	}
	return routine, nil
}
