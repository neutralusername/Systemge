package serviceSingleRequest

import (
	"errors"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func NewSync[D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterConfig *configs.Accepter,
	readerConfig *configs.ReaderSync,
	routineConfig *configs.Routine,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	readHandler tools.ReadHandlerWithResult[D, systemge.Connection[D]],
) (*SingleRequestServer[D], error) {

	if readHandler == nil {
		return nil, errors.New("readerHandler is nil")
	}
	if acceptHandler == nil {
		return nil, errors.New("acceptHandler is nil")
	}
	if readerConfig == nil {
		return nil, errors.New("readerConfig is nil")
	}

	singleReuqestSync := &SingleRequestServer[D]{}

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
			case <-singleReuqestSync.GetAccepter().GetRoutine().GetStopChannel():
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
