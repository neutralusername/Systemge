package serviceSingleRequest

import (
	"github.com/neutralusername/systemge/configs"
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
		accepterConfig,
		routineConfig,
		listener,
		func(connection systemge.Connection[D]) error {
			if err := acceptHandler(connection); err != nil {
				// do smthg with the error
				return err
			}

			// could use a serviceReader for this singular request but this would come with some avoidable overhead

			data, err := connection.Read(readerConfig.ReadTimeoutNs) // fix this blocking call and make it dependent on the accepter stop channel / close connection
			if err != nil {
				// do smthg with the error
				return err
			}
			result, err := readHandler(data, connection)
			if err != nil {
				// do smthg with the error
				return err
			}

			if err := connection.Write(result, readerConfig.WriteTimeoutNs); err != nil { // fix this blocking call and make it dependent on the accepter stop channel / close connection
				// do smthg with the error
				return err
			}

			connection.Close()
			return nil
		},
		handleRequestsConcurrently,
	)
}
