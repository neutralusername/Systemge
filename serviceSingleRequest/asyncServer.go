package serviceSingleRequest

import (
	"errors"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func NewAsync[D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterConfig *configs.AccepterServer,
	readerConfig *configs.ReaderServerAsync,
	routineConfig *configs.Routine,
	handleRequestsConcurrently bool,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	readHandler tools.ReadHandler[D, systemge.Connection[D]],
) (*serviceAccepter.Accepter[D], error) {

	return serviceAccepter.NewAccepterServer(
		listener,
		accepterConfig,
		routineConfig,
		handleRequestsConcurrently,
		func(stopChannel <-chan struct{}, connection systemge.Connection[D]) error {
			if err := acceptHandler(stopChannel, connection); err != nil {
				// do smthg with the error
				return err
			}

			select {
			case <-stopChannel:
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
				readHandler(stopChannel, data, connection)
				connection.Close()
				return nil
			}
		},
	)
}
