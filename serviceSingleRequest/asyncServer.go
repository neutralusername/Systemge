package serviceSingleRequest

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func NewSingleRequestServerAsync[D any](
	accepterConfig *configs.AccepterServer,
	readerConfig *configs.ReaderServerAsync,
	routineConfig *configs.Routine,
	listener systemge.Listener[D, systemge.Connection[D]],
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	readHandler tools.ReadHandler[D, systemge.Connection[D]],
	handleAcceptsConcurrently bool,
) (*serviceAccepter.AccepterServer[D], error) {

	return serviceAccepter.NewAccepterServer(
		accepterConfig,
		routineConfig,
		listener,
		func(connection systemge.Connection[D]) error {
			if err := acceptHandler(connection); err != nil {
				// do smthg with the error
				return err
			}

			data, err := connection.Read(readerConfig.ReadTimeoutNs) // fix this blocking call and make it dependent on the accepter stop channel / close connection
			if err != nil {
				// do smthg with the error
				return err
			}
			readHandler(data, connection)
			connection.Close()
			return nil
		},
		handleAcceptsConcurrently,
	)
}
