package serviceTypedListener

import (
	"errors"

	"github.com/neutralusername/systemge/serviceTypedConnection"
	"github.com/neutralusername/systemge/systemge"
)

type typedListener[O any, D any] struct {
	systemge.Listener[D, systemge.Connection[D]]
	deserializer func(D) (O, error)
	serializer   func(O) (D, error)
}

func New[O any, D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	deserializer func(D) (O, error),
	serializer func(O) (D, error),
) (systemge.Listener[O, systemge.Connection[O]], error) {

	if listener == nil {
		return nil, errors.New("connection is nil")
	}
	if serializer == nil {
		return nil, errors.New("serializer is nil")
	}
	if deserializer == nil {
		return nil, errors.New("deserializer is nil")
	}

	typedListener := &typedListener[O, D]{
		Listener:     listener,
		deserializer: deserializer,
		serializer:   serializer,
	}
	return typedListener, nil
}

func (typedListener *typedListener[O, D]) Accept(timeoutNs int64) (systemge.Connection[O], error) {
	connection, err := typedListener.Listener.Accept(timeoutNs)
	if err != nil {
		return nil, err
	}

	return serviceTypedConnection.New(
		connection,
		typedListener.deserializer,
		typedListener.serializer,
	)
}

func (typedListener *typedListener[O, D]) GetConnector() systemge.Connector[O, systemge.Connection[O]] {
	return nil
}
