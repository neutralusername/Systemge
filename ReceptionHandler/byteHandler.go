package ReceptionHandler

import "github.com/neutralusername/Systemge/Tools"

type ByteHandler[T any] func([]byte) error

func NewChainByteHandler[T any](
	handlers ...ByteHandler[T],
) ByteHandler[T] {
	return func(bytes []byte) error {
		for _, handler := range handlers {
			if err := handler(bytes); err != nil {
				return err
			}
		}
		return nil
	}
}

func NewByteRateLimitByteHandler[T any](
	tokenBucketRateLimiter *Tools.TokenBucketRateLimiter,
) ByteHandler[T] {
	return func(bytes []byte) error {
		tokenBucketRateLimiter.Consume(uint64(len(bytes)))
		return nil
	}
}

func NewMessageRateLimitByteHandler[T any](
	tokenBucketRateLimiter *Tools.TokenBucketRateLimiter,
) ByteHandler[T] {
	return func(bytes []byte) error {
		tokenBucketRateLimiter.Consume(1)
		return nil
	}
}
