package tools

type ReadHandler[D any, C any] func(<-chan struct{}, D, C)

type ReadHandlerWithResult[D any, C any] func(<-chan struct{}, D, C) (D, error)

type ReadHandlerWithError[D any, C any] func(<-chan struct{}, D, C) error

/*
type ByteHandler[D any, C any] func(D, C) error
type ObjectDeserializer[D any, O any, C any] func(D, C) (O, error)
type ObjectHandler[O any, C any] func(O, C) error

type ReadHandlerQueueWrapper[O any, C any] struct {
	object O
	caller C
}

func NewDefaultReadHandler[D any, O any, C any](
	byteHandler ByteHandler[D, C],
	deserializer ObjectDeserializer[D, O, C],
	objectHandler ObjectHandler[O, C],
) ReadHandler[D, C] {
	return func(bytes D, caller C) {

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
func NewByteHandler[D any, C any](handlers ...ByteHandler[D, C]) ByteHandler[D, C] {
	return func(bytes D, caller C) error {
		for _, handler := range handlers {
			if err := handler(bytes, caller); err != nil {
				return err
			}
		}
		return nil
	}
}

type ObtainTokensFromBytes[D any] func(D) uint64

func NewTokenBucketRateLimitHandler[D any, C any](obtainTokensFromBytes ObtainTokensFromBytes[D], tokenBucketRateLimiterConfig *Config.TokenBucketRateLimiter) ByteHandler[D, C] {
	tokenBucketRateLimiter := NewTokenBucketRateLimiter(tokenBucketRateLimiterConfig)
	return func(bytes D, caller C) error {
		tokens := obtainTokensFromBytes(bytes)
		tokenBucketRateLimiter.Consume(tokens)
		return nil
	}
}
*/
