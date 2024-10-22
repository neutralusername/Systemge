package typedConnection

import (
	"errors"

	"github.com/neutralusername/systemge/systemge"
)

type typedConnection[O any, D any] struct {
	systemge.Connection[D]
	deserializer func(D) (O, error)
	serializer   func(O) (D, error)
}

func New[O any, D any](
	connection systemge.Connection[D],
	deserializer func(D) (O, error),
	serializer func(O) (D, error),
) (systemge.Connection[O], error) {

	if connection == nil {
		return nil, errors.New("connection is nil")
	}
	if serializer == nil {
		return nil, errors.New("serializer is nil")
	}
	if deserializer == nil {
		return nil, errors.New("deserializer is nil")
	}

	typedConnection := &typedConnection[O, D]{
		Connection:   connection,
		deserializer: deserializer,
		serializer:   serializer,
	}
	return typedConnection, nil
}

func (typedConnection *typedConnection[O, D]) Read(timeoutNs int64) (O, error) {
	data, err := typedConnection.Connection.Read(timeoutNs)
	if err != nil {
		var nilValue O
		return nilValue, err
	}
	return typedConnection.deserializer(data)
}

func (typedConnection *typedConnection[O, D]) Write(object O, timeoutNs int64) error {
	serializedData, err := typedConnection.serializer(object)
	if err != nil {
		return err
	}
	return typedConnection.Connection.Write(serializedData, timeoutNs)
}
