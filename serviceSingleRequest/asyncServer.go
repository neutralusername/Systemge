package serviceSingleRequest

import (
	"errors"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type SingleReuqest[D any] struct {
	accepter *serviceAccepter.Accepter[D]
}

func (s *SingleReuqest[D]) GetAccepter() *serviceAccepter.Accepter[D] {
	return s.accepter
}

func NewAsync[D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterConfig *configs.AccepterServer,
	readerConfig *configs.ReaderServerAsync,
	routineConfig *configs.Routine,
	handleRequestsConcurrently bool,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	readHandler tools.ReadHandler[D, systemge.Connection[D]],
) (*serviceAccepter.Accepter[D], error) {

	singleReuqestAsync := &SingleReuqest[D]{}

	accepter, err := serviceAccepter.NewAccepterServer(
		listener,
		accepterConfig,
		routineConfig,
		handleRequestsConcurrently,
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
