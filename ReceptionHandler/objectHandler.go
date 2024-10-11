package ReceptionHandler

import "github.com/neutralusername/Systemge/Tools"

type ObjectHandler[T any] func(T) error

type ObtainEnqueueConfigs[T any] func(T) (token string, priority uint32, timeout uint32)
type ObtainResponseToken[T any] func(T) string
type ObjectValidator[T any] func(T) error
type ObtainTopic[T any] func(T) string

// executes all handlers in order, return error if any handler returns an error
func NewChainObjecthandler[T any](
	handlers ...ObjectHandler[T],
) ObjectHandler[T] {
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
	priorityTokenQueue *Tools.PriorityTokenQueue[T],
	obtainEnqueueConfigs ObtainEnqueueConfigs[T],
) ObjectHandler[T] {
	return func(object T) error {
		token, priority, timeoutMs := obtainEnqueueConfigs(object)
		return priorityTokenQueue.Push(token, object, priority, timeoutMs)
	}
}

func NewResponseObjectHandler[T any](
	requestResponseManager *Tools.RequestResponseManager[T],
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

func NewTopicObjectHandler[T any](
	topicManager *Tools.TopicManager,
	obtainTopic func(T) string,
) ObjectHandler[T] {
	return func(object T) error {
		if topicManager != nil {
			result, err := topicManager.Handle(obtainTopic(object), object)

		}
		return nil
	}
}

func NewValidationObjectHandler[T any](
	validator ObjectValidator[T],
) ObjectHandler[T] {
	return func(object T) error {
		return validator(object)
	}
}
