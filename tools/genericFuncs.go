package tools

type Serializer[T any, O any] func(object O) (T, error)
type Deserializer[T any, O any] func(data T) (O, error)
