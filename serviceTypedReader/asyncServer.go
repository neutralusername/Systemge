package serviceTypedReader

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceReader"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func NewTypesReaderAsync[D any, O any](
	connection systemge.Connection[D],
	readerServerAsyncConfig *configs.ReaderAsync,
	routineConfig *configs.Routine,
	readHandler tools.ReadHandler[O, systemge.Connection[D]],
	deserializer func(D) (O, error),
) (*serviceReader.ReaderAsync[D], error) {
	reader, err := serviceReader.NewAsync(
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
	return reader, err
}
