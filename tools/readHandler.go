package tools

type ReadHandlerFactory[B any, C any] func() ReadHandler[B, C]

type ReadHandler[B any, C any] func(B, C)

type ReadHandlerWithResult[B any, C any] func(B, C) B

type ReadHandlerWithError[B any, C any] func(B, C) error

/*
type ByteHandler[B any, C any] func(B, C) error
type ObjectDeserializer[B any, O any, C any] func(B, C) (O, error)
type ObjectHandler[O any, C any] func(O, C) error

type ReadHandlerQueueWrapper[O any, C any] struct {
	object O
	caller C
}

func NewDefaultReadHandler[B any, O any, C any](
	byteHandler ByteHandler[B, C],
	deserializer ObjectDeserializer[B, O, C],
	objectHandler ObjectHandler[O, C],
) ReadHandler[B, C] {
	return func(bytes B, caller C) {

		err := byteHandler(bytes, caller)
		if err != nil {
			return
		}

		object, err := deserializer(bytes, caller)
		if err != nil {
			return
		}

		objectHandler(object, caller)
	}
}

// executes all handlers in order, return error if any handler returns an error
func NewObjectHandler[O any, C any](handlers ...ObjectHandler[O, C]) ObjectHandler[O, C] {
	return func(object O, caller C) error {
		for _, handler := range handlers {
			if err := handler(object, caller); err != nil {
				return err
			}
		}
		return nil
	}
}

type ObtainReadHandlerEnqueueConfigs[O any, C any] func(O, C) (token string, priority uint32, timeout uint32)

func NewQueueObjectHandler[O any, C any](
	priorityTokenQueue *PriorityTokenQueue[*ReadHandlerQueueWrapper[O, C]],
	obtainEnqueueConfigs ObtainReadHandlerEnqueueConfigs[O, C],
) ObjectHandler[O, C] {
	return func(object O, caller C) error {
		token, priority, timeoutMs := obtainEnqueueConfigs(object, caller)
		queueWrapper := &ReadHandlerQueueWrapper[O, C]{
			object,
			caller,
		}
		return priorityTokenQueue.Push(token, queueWrapper, priority, timeoutMs)
	}
}

type ObjectValidator[O any, C any] func(O, C) error

func NewValidationObjectHandler[O any, C any](validator ObjectValidator[O, C]) ObjectHandler[O, C] {
	return func(object O, caller C) error {
		return validator(object, caller)
	}
}

// executes all handlers in order, return error if any handler returns an error
func NewByteHandler[B any, C any](handlers ...ByteHandler[B, C]) ByteHandler[B, C] {
	return func(bytes B, caller C) error {
		for _, handler := range handlers {
			if err := handler(bytes, caller); err != nil {
				return err
			}
		}
		return nil
	}
}

type ObtainTokensFromBytes[B any] func(B) uint64

func NewTokenBucketRateLimitHandler[B any, C any](obtainTokensFromBytes ObtainTokensFromBytes[B], tokenBucketRateLimiterConfig *Config.TokenBucketRateLimiter) ByteHandler[B, C] {
	tokenBucketRateLimiter := NewTokenBucketRateLimiter(tokenBucketRateLimiterConfig)
	return func(bytes B, caller C) error {
		tokens := obtainTokensFromBytes(bytes)
		tokenBucketRateLimiter.Consume(tokens)
		return nil
	}
}
*/
