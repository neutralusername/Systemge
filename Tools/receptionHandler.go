package Tools

type ReceptionHandlerFactory[S any] func() ReceptionHandler[S]
type ReceptionHandler[S any] func([]byte, S) error
type ObjectDeserializer[T any, S any] func([]byte, S) (T, error)
type ByteHandler[S any] func([]byte, S) error

func NewReceptionHandler[T any, S any](
	byteHandler ByteHandler[S],
	deserializer ObjectDeserializer[T, S],
	objectHandler ObjectHandler[T, S],
) ReceptionHandler[S] {
	return func(bytes []byte, structName123 S) error {

		err := byteHandler(bytes, structName123)
		if err != nil {
			return err
		}

		object, err := deserializer(bytes, structName123)
		if err != nil {
			return err
		}

		return objectHandler(object, structName123)
	}
}

type ObjectHandler[T any, S any] func(T, S) error

type ObtainEnqueueConfigs[T any] func(T) (token string, priority uint32, timeout uint32)
type ObtainResponseToken[T any] func(T) string
type ObjectValidator[T any, S any] func(T, S) error
type ObtainTopic[T any] func(T) string
type ResultHandler[T any] func(T) error

// executes all handlers in order, return error if any handler returns an error
func NewChainObjecthandler[T any, S any](handlers ...ObjectHandler[T, S]) ObjectHandler[T, S] {
	return func(object T, structName123 S) error {
		for _, handler := range handlers {
			if err := handler(object, structName123); err != nil {
				return err
			}
		}
		return nil
	}
}

func NewQueueObjectHandler[T any, S any](
	priorityTokenQueue *PriorityTokenQueue[T],
	obtainEnqueueConfigs ObtainEnqueueConfigs[T],
) ObjectHandler[T, S] {
	return func(object T, structName123 S) error {
		token, priority, timeoutMs := obtainEnqueueConfigs(object)
		return priorityTokenQueue.Push(token, object, priority, timeoutMs)
	}
}

func NewResponseObjectHandler[T any, S any](
	requestResponseManager *RequestResponseManager[T],
	obtainResponseToken ObtainResponseToken[T],
) ObjectHandler[T, S] {
	return func(object T, structName123 S) error {
		responseToken := obtainResponseToken(object)
		if responseToken != "" {
			if requestResponseManager != nil {
				return requestResponseManager.AddResponse(responseToken, object)
			}
		}
		return nil
	}
}

// resultHandler requires check for nil if applicable
func NewTopicObjectHandler[T any, R any, S any](
	topicManager *TopicManager[T, R],
	obtainTopic func(T) string,
	resultHandler ResultHandler[R],
) ObjectHandler[T, S] {
	return func(object T, structName123 S) error {
		if topicManager != nil {
			result, err := topicManager.Handle(obtainTopic(object), object)
			if err != nil {
				return err
			}
			return resultHandler(result)
		}
		return nil
	}
}

func NewValidationObjectHandler[T any, S any](validator ObjectValidator[T, S]) ObjectHandler[T, S] {
	return func(object T, structName123 S) error {
		return validator(object, structName123)
	}
}

// executes all handlers in order, return error if any handler returns an error
func NewChainByteHandler[S any](handlers ...ByteHandler[S]) ByteHandler[S] {
	return func(bytes []byte, structName123 S) error {
		for _, handler := range handlers {
			if err := handler(bytes, structName123); err != nil {
				return err
			}
		}
		return nil
	}
}

func NewByteRateLimitByteHandler[S any](tokenBucketRateLimiter *TokenBucketRateLimiter, structName123 S) ByteHandler[S] {
	return func(bytes []byte, structName123 S) error {
		tokenBucketRateLimiter.Consume(uint64(len(bytes)))
		return nil
	}
}

func NewMessageRateLimitByteHandler[S any](tokenBucketRateLimiter *TokenBucketRateLimiter, structName123 S) ByteHandler[S] {
	return func(bytes []byte, structName123 S) error {
		tokenBucketRateLimiter.Consume(1)
		return nil
	}
}