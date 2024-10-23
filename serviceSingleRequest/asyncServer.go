package serviceSingleRequest

import (
	"errors"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type SingleRequestServer[D any] struct {
	accepter *serviceAccepter.Accepter[D]
}

func (s *SingleRequestServer[D]) GetAccepter() *serviceAccepter.Accepter[D] {
	return s.accepter
}

func NewAsync[D any](
	listener systemge.Listener[D],
	accepterConfig *configs.Accepter,
	readerConfig *configs.ReaderAsync,
	routineConfig *configs.Routine,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	readHandler tools.ReadHandler[D, systemge.Connection[D]],
) (*serviceAccepter.Accepter[D], error) {

	if readHandler == nil {
		return nil, errors.New("readerHandler is nil")
	}
	if acceptHandler == nil {
		return nil, errors.New("acceptHandler is nil")
	}
	if readerConfig == nil {
		return nil, errors.New("readerConfig is nil")
	}

	singleReuqestAsync := &SingleRequestServer[D]{}

	accepter, err := serviceAccepter.New(
		listener,
		accepterConfig,
		routineConfig,
		func(connection systemge.Connection[D]) error {
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

			case data, ok := <-helpers.ChannelCall(func() (D, error) { return connection.Read(readerConfig.ReadTimeoutNs) }):
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
