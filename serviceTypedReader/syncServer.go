package serviceTypedReader

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/serviceReader"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func NewTypedReaderSync[D any, O any](
	connection systemge.Connection[D],
	readerServerSyncConfig *configs.ReaderSync,
	routineConfig *configs.Routine,
	readHandler tools.ReadHandler[O, systemge.Connection[D]],
	deserializer func(D) (O, error),
	serializer func(O) (D, error),
) (*serviceReader.ReaderSync[D], error) {

	reader, err := serviceReader.NewSync(
		connection,
		readerServerSyncConfig,
		routineConfig,
		func(data D, connection systemge.Connection[D]) (D, error) {
			object, err := deserializer(data)
			if err != nil {
				return helpers.GetNilValue(data), err
			}
			data, err = serializer(object)
			if err != nil {
				return helpers.GetNilValue(data), err
			}
			return data, nil
		},
	)
	return reader, err
}
