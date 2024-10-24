package typedConnection

import (
	"errors"

	"github.com/neutralusername/systemge/systemge"
)

type typedConnection[T any, O any] struct {
	systemge.Connection[T]
	serializer   func(O) (T, error)
	deserializer func(T) (O, error)
}

func New[T any, O any](
	connection systemge.Connection[T],
	serializer func(O) (T, error),
	deserializer func(T) (O, error),
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

	typedConnection := &typedConnection[T, O]{
		Connection:   connection,
		serializer:   serializer,
		deserializer: deserializer,
	}
	return typedConnection, nil
}

func (typedConnection *typedConnection[T, O]) Read(timeoutNs int64) (O, error) {
	data, err := typedConnection.Connection.Read(timeoutNs)
	if err != nil {
		var nilValue O
		return nilValue, err
	}
	return typedConnection.deserializer(data)
}

func (typedConnection *typedConnection[T, O]) Write(data O, timeoutNs int64) error {
	serializedData, err := typedConnection.serializer(data)
	if err != nil {
		return err
	}
	return typedConnection.Connection.Write(serializedData, timeoutNs)
}
