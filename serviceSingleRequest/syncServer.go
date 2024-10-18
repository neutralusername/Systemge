package serviceSingleRequest

import (
	"errors"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func NewSingleRequestServerSync[D any](
	accepterConfig *configs.AccepterServer,
	readerConfig *configs.ReaderServerSync,
	routineConfig *configs.Routine,
	listener systemge.Listener[D, systemge.Connection[D]],
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	readHandler tools.ReadHandlerWithResult[D, systemge.Connection[D]],
	handleRequestsConcurrently bool,
) (*serviceAccepter.AccepterServer[D], error) {

	return serviceAccepter.NewAccepterServer(
		listener,
		accepterConfig,
		routineConfig,
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

				result, err := readHandler(stopChannel, data, connection)
				if err != nil {
					// do smthg with the error
					return err
				}

				select {
				case <-stopChannel:
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
		handleRequestsConcurrently,
	)
}
