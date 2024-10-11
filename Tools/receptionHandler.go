package Tools

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

type ObjectHandler[T any] func(T) error

type ObtainEnqueueConfigs[T any] func(T) (token string, priority uint32, timeout uint32)
type ObtainResponseToken[T any] func(T) string
type ObjectValidator[T any] func(T) error
type ObtainTopic[T any] func(T) string
type ResultHandler[T any] func(T) error

// executes all handlers in order, return error if any handler returns an error
func NewChainObjecthandler[T any](handlers ...ObjectHandler[T]) ObjectHandler[T] {
	return func(object T) error {
		for _, handler := range handlers {
			if err := handler(object); err != nil {
				return err
			}
		}
		return nil
	}
}

func NewQueueObjectHandler[T any](
	priorityTokenQueue *PriorityTokenQueue[T],
	obtainEnqueueConfigs ObtainEnqueueConfigs[T],
) ObjectHandler[T] {
	return func(object T) error {
		token, priority, timeoutMs := obtainEnqueueConfigs(object)
		return priorityTokenQueue.Push(token, object, priority, timeoutMs)
	}
}

func NewResponseObjectHandler[T any](
	requestResponseManager *RequestResponseManager[T],
	obtainResponseToken ObtainResponseToken[T],
) ObjectHandler[T] {
	return func(object T) error {
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
func NewTopicObjectHandler[P any, R any](
	topicManager *TopicManager[P, R],
	obtainTopic func(P) string,
	resultHandler ResultHandler[R],
) ObjectHandler[P] {
	return func(object P) error {
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

func NewValidationObjectHandler[T any](validator ObjectValidator[T]) ObjectHandler[T] {
	return func(object T) error {
		return validator(object)
	}
}

type ByteHandler[T any] func([]byte) error

// executes all handlers in order, return error if any handler returns an error
func NewChainByteHandler[T any](handlers ...ByteHandler[T]) ByteHandler[T] {
	return func(bytes []byte) error {
		for _, handler := range handlers {
			if err := handler(bytes); err != nil {
				return err
			}
		}
		return nil
	}
}

func NewByteRateLimitByteHandler[T any](tokenBucketRateLimiter *TokenBucketRateLimiter) ByteHandler[T] {
	return func(bytes []byte) error {
		tokenBucketRateLimiter.Consume(uint64(len(bytes)))
		return nil
	}
}

func NewMessageRateLimitByteHandler[T any](tokenBucketRateLimiter *TokenBucketRateLimiter) ByteHandler[T] {
	return func(bytes []byte) error {
		tokenBucketRateLimiter.Consume(1)
		return nil
	}
}
