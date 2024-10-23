package serviceSingleRequest

import (
	"errors"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/systemge"
)

func NewSync[T any](
	listener systemge.Listener[T],
	accepterConfig *configs.Accepter,
	readerConfig *configs.ReaderSync,
	routineConfig *configs.Routine,
	acceptHandler systemge.AcceptHandlerWithError[T],
	readHandler systemge.ReadHandlerWithResult[T],
) (*SingleRequestServer[T], error) {

	if readHandler == nil {
		return nil, errors.New("readerHandler is nil")
	}
	if acceptHandler == nil {
		return nil, errors.New("acceptHandler is nil")
	}
	if readerConfig == nil {
		return nil, errors.New("readerConfig is nil")
	}

	singleReuqestSync := &SingleRequestServer[T]{}

	accepter, err := serviceAccepter.New(
		listener,
		accepterConfig,
		routineConfig,
		func(connection systemge.Connection[T]) error {
			if err := acceptHandler(connection); err != nil {
				// do smthg with the error
				return err
			}

			select {
			case <-singleReuqestSync.GetAccepter().GetRoutine().GetStopChannel():
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
				if err != nil {
					// do smthg with the error
					return err
				}

				select {
				case <-singleReuqestSync.GetAccepter().GetRoutine().GetStopChannel():
					connection.SetWriteDeadline(1)
					// routine was stopped
					return errors.New("routine was stopped")

				case <-connection.GetCloseChannel():
					// ending routine due to connection close
					return errors.New("connection was closed")

				case <-helpers.ChannelCall(func() (error, error) {
					if err := connection.Write(result, readerConfig.WriteTimeoutNs); err != nil {
						// do smthg with the error
						return err, err
					}
					return nil, nil
				}):
					connection.Close()
					return nil
				}
			}
		},
	)
	if err != nil {
		return nil, err
	}
	singleReuqestSync.accepter = accepter

	return singleReuqestSync, nil
}
