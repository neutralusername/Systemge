package serviceTypedAccepter

import (
	"errors"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/serviceTypedConnection"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func New[D any, O any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterConfig *configs.Accepter,
	routineConfig *configs.Routine,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[O]],
	serializer func(D) (O, error),
	deserializer func(O) (D, error),
) (*serviceAccepter.Accepter[D], error) {

	if serializer == nil {
		return nil, errors.New("serializer is nil")
	}
	if deserializer == nil {
		return nil, errors.New("deserializer is nil")
	}
	if acceptHandler == nil {
		return nil, errors.New("acceptHandler is nil")
	}

	return serviceAccepter.New[D](
		listener,
		accepterConfig,
		routineConfig,
		func(connection systemge.Connection[D]) error {
			typedConnection, err := serviceTypedConnection.New[O, D](
				connection,
				serializer,
				deserializer,
			)
			if err != nil {
				return err
			}
			return acceptHandler(typedConnection)
		},
	)
}
