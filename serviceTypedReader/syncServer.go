package serviceTypedReader

import (
	"errors"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/serviceReader"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func NewSync[D any, O any](
	connection systemge.Connection[D],
	readerServerSyncConfig *configs.ReaderSync,
	routineConfig *configs.Routine,
	readHandler tools.ReadHandlerWithResult[O, systemge.Connection[D]],
	deserializer func(D) (O, error),
	serializer func(O) (D, error),
) (*serviceReader.ReaderSync[D], error) {

	if deserializer == nil {
		return nil, errors.New("deserializer is nil")
	}
	if serializer == nil {
		return nil, errors.New("serializer is nil")
	}
	if readHandler == nil {
		return nil, errors.New("readHandler is nil")
	}

	return serviceReader.NewSync(
		connection,
		readerServerSyncConfig,
		routineConfig,
		func(data D, connection systemge.Connection[D]) (D, error) {
			object, err := deserializer(data)
			if err != nil {
				return helpers.GetNilValue(data), err
			}

			resultObject, err := readHandler(object, connection)
			if err != nil {
				return helpers.GetNilValue(data), err
			}

			data, err = serializer(resultObject)
			if err != nil {
				return helpers.GetNilValue(data), err
			}
			return data, nil
		},
	)
}
