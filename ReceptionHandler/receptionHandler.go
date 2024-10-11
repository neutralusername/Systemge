package ReceptionHandler

type ReceptionHandler func([]byte) error
type ObjectDeserializer[T any] func([]byte) (T, error)

func NewReceptionHandler[T any](
	byteHandler ByteHandler[T],
	deserializer ObjectDeserializer[T],
	objectHandler ObjectHandler[T],
) ReceptionHandler {
	return func(bytes []byte) error {

		err := byteHandler(bytes)
		if err != nil {
			return err
		}

		object, err := deserializer(bytes)
		if err != nil {
			return err
		}

		return objectHandler(object)
	}
}
