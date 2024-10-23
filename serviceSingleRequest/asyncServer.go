package serviceSingleRequest

import (
	"errors"

	"github.com/neutralusername/systemge/accepter"
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/systemge"
)

type SingleRequestServer[T any] struct {
	accepter *accepter.Accepter[T]
}

func (s *SingleRequestServer[T]) GetAccepter() *accepter.Accepter[T] {
	return s.accepter
}

func NewAsync[T any](
	listener systemge.Listener[T],
	accepterConfig *configs.Accepter,
	readerConfig *configs.ReaderAsync,
	routineConfig *configs.Routine,
	acceptHandler systemge.AcceptHandlerWithError[T],
	readHandler systemge.ReadHandler[T],
) (*accepter.Accepter[T], error) {

	if readHandler == nil {
		return nil, errors.New("readerHandler is nil")
	}
	if acceptHandler == nil {
		return nil, errors.New("acceptHandler is nil")
	}
	if readerConfig == nil {
		return nil, errors.New("readerConfig is nil")
	}

	singleReuqestAsync := &SingleRequestServer[T]{}

	accepter, err := accepter.New(
		listener,
		accepterConfig,
		routineConfig,
		func(connection systemge.Connection[T]) error {
			if err := acceptHandler(connection); err != nil {
				// do smthg with the error
				return err
			}

			select {
			case <-singleReuqestAsync.GetAccepter().GetRoutine().GetStopChannel():
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
		},
	)
	if err != nil {
		return nil, err
	}
	singleReuqestAsync.accepter = accepter

	return accepter, nil
}
