package serviceTypedConnection

import (
	"errors"

	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type typedConnection[O any, D any] struct {
	connection   systemge.Connection[D]
	serializer   func(D) (O, error)
	deserializer func(O) (D, error)
}

func New[O any, D any](
	connection systemge.Connection[D],
	serializer func(D) (O, error),
	deserializer func(O) (D, error),
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
		connection: connection,
	}
	return typedConnection, nil
}

func (c *typedConnection[O, D]) Close() error {
	return c.connection.Close()
}

func (c *typedConnection[O, D]) GetInstanceId() string {
	return c.connection.GetInstanceId()
}

func (c *typedConnection[O, D]) GetAddress() string {
	return c.connection.GetAddress()
}

func (c *typedConnection[O, D]) GetStatus() int {
	return c.connection.GetStatus()
}

func (c *typedConnection[O, D]) GetCloseChannel() <-chan struct{} {
	return c.connection.GetCloseChannel()
}

func (c *typedConnection[O, D]) Read(timeoutNs int64) (O, error) {
	data, err := c.connection.Read(timeoutNs)
	if err != nil {
		var nilValue O
		return nilValue, err
	}
	return c.serializer(data)
}

func (c *typedConnection[O, D]) SetReadDeadline(timeoutNs int64) {
	c.connection.SetReadDeadline(timeoutNs)
}

func (c *typedConnection[O, D]) Write(data O, timeoutNs int64) error {
	serializedData, err := c.deserializer(data)
	if err != nil {
		return err
	}
	return c.connection.Write(serializedData, timeoutNs)
}

func (c *typedConnection[O, D]) SetWriteDeadline(timeoutNs int64) {
	c.connection.SetWriteDeadline(timeoutNs)
}

func (c *typedConnection[O, D]) GetDefaultCommands() tools.CommandHandlers {
	return c.connection.GetDefaultCommands()
}

func (c *typedConnection[O, D]) GetMetrics() tools.MetricsTypes {
	return c.connection.GetMetrics()
}

func (c *typedConnection[O, D]) CheckMetrics() tools.MetricsTypes {
	return c.connection.CheckMetrics()
}
