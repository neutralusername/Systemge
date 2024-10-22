package serviceTypedReader

import (
	"errors"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceReader"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func NewAsync[D any, O any](
	connection systemge.Connection[D],
	readerServerAsyncConfig *configs.ReaderAsync,
	routineConfig *configs.Routine,
	readHandler tools.ReadHandler[O, systemge.Connection[D]],
	deserializer func(D) (O, error),
) (*serviceReader.ReaderAsync[D], error) {

	if deserializer == nil {
		return nil, errors.New("deserializer is nil")
	}
	if readHandler == nil {
		return nil, errors.New("readHandler is nil")
	}

	return serviceReader.NewAsync(
		connection,
		readerServerAsyncConfig,
		routineConfig,
		func(data D, connection systemge.Connection[D]) {
			object, err := deserializer(data)
			if err != nil {
				return
			}
			readHandler(object, connection)
		},
	)
}
