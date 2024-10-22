package serviceTypedConnection

import (
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type typedConnection[O any] struct {
	connection   systemge.Connection[[]byte]
	serializer   func([]byte) (O, error)
	deserializer func(O) ([]byte, error)
}

func New[O any](connection systemge.Connection[[]byte]) (systemge.Connection[O], error) {
	typedConnection := &typedConnection[O]{
		connection: connection,
	}
	return typedConnection, nil
}

func (c *typedConnection[O]) Close() error {
	return c.connection.Close()
}

func (c *typedConnection[O]) GetInstanceId() string {
	return c.connection.GetInstanceId()
}

func (c *typedConnection[O]) GetAddress() string {
	return c.connection.GetAddress()
}

func (c *typedConnection[O]) GetStatus() int {
	return c.connection.GetStatus()
}

func (c *typedConnection[O]) GetCloseChannel() <-chan struct{} {
	return c.connection.GetCloseChannel()
}

func (c *typedConnection[O]) Read(timeoutNs int64) (O, error) {
	data, err := c.connection.Read(timeoutNs)
	if err != nil {
		var nilValue O
		return nilValue, err
	}
	return c.serializer(data)
}

func (c *typedConnection[O]) SetReadDeadline(timeoutNs int64) {
	c.connection.SetReadDeadline(timeoutNs)
}

func (c *typedConnection[O]) Write(data O, timeoutNs int64) error {
	serializedData, err := c.deserializer(data)
	if err != nil {
		return err
	}
	return c.connection.Write(serializedData, timeoutNs)
}

func (c *typedConnection[O]) SetWriteDeadline(timeoutNs int64) {
	c.connection.SetWriteDeadline(timeoutNs)
}

func (c *typedConnection[O]) GetDefaultCommands() tools.CommandHandlers {
	return c.connection.GetDefaultCommands()
}

func (c *typedConnection[O]) GetMetrics() tools.MetricsTypes {
	return c.connection.GetMetrics()
}

func (c *typedConnection[O]) CheckMetrics() tools.MetricsTypes {
	return c.connection.CheckMetrics()
}
