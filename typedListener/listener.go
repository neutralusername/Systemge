package typedListener

import (
	"errors"

	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/typedConnection"
)

type typedListener[T any, O any] struct {
	systemge.Listener[T]
	serializer   func(O) (T, error)
	deserializer func(T) (O, error)
}

func New[T any, O any](
	listener systemge.Listener[T],
	serializer func(O) (T, error),
	deserializer func(T) (O, error),
) (systemge.Listener[O], error) {

	if listener == nil {
		return nil, errors.New("connection is nil")
	}
	if serializer == nil {
		return nil, errors.New("serializer is nil")
	}
	if deserializer == nil {
		return nil, errors.New("deserializer is nil")
	}

	typedListener := &typedListener[T, O]{
		Listener:     listener,
		serializer:   serializer,
		deserializer: deserializer,
	}
	return typedListener, nil
}

func (typedListener *typedListener[T, O]) Accept(timeoutNs int64) (systemge.Connection[O], error) {
	connection, err := typedListener.Listener.Accept(timeoutNs)
	if err != nil {
		return nil, err
	}

	return typedConnection.New(
		connection,
		typedListener.deserializer,
		typedListener.serializer,
	)
}

type connector[T any, O any] struct {
	systemge.Connector[T]
	deserializer func(T) (O, error)
	serializer   func(O) (T, error)
}

func (typedListener *typedListener[T, O]) GetConnector() systemge.Connector[O] {
	return &connector[T, O]{
		typedListener.Listener.GetConnector(),
		typedListener.deserializer,
		typedListener.serializer,
	}
}

func (connector *connector[T, O]) Connect(timeout int64) (systemge.Connection[O], error) {
	connection, err := connector.Connector.Connect(timeout)
	if err != nil {
		return nil, err
	}

	return typedConnection.New(
		connection,
		connector.deserializer,
		connector.serializer,
	)
}
