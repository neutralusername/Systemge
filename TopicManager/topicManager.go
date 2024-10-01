package TopicManager

import (
	"errors"
	"sync"
)

type TopicHandler func(...any) (any, error)
type TopicHandlers map[string]TopicHandler

type TopicManager struct {
	topicHandlers       TopicHandlers
	unknownTopicHandler TopicHandler

	isCLosed bool

	queue             chan *queueStruct
	topicQueues       map[string]chan *queueStruct
	topicStopChannels map[string]chan struct{}
	unknownTopicQueue chan *queueStruct

	mutex sync.Mutex

	queueSize       uint32
	topicQueueSize  uint32
	concurrentCalls bool
}

type queueStruct struct {
	topic                string
	args                 []any
	responseAnyChannel   chan any
	responseErrorChannel chan error
}

// modes: (l == large enough to never be full)
// topicQueueSize: 0, queueSize: l concurrentCalls: false -> "sequential"
// topicQueueSize: l, queueSize: l concurrentCalls: false -> "topic exclusive"
// topicQueueSize: 0|l, queueSize: 0|l concurrentCalls: true -> "concurrent"

func NewTopicManager(topicHandlers TopicHandlers, unknownTopicHandler TopicHandler, topicQueueSize uint32, queueSize uint32, concurrentCalls bool) *TopicManager {
	if topicHandlers == nil {
		topicHandlers = make(TopicHandlers)
	}
	topicManager := &TopicManager{
		topicHandlers:       topicHandlers,
		unknownTopicHandler: unknownTopicHandler,
		queue:               make(chan *queueStruct, queueSize),
		topicQueues:         make(map[string]chan *queueStruct),
		topicStopChannels:   make(map[string]chan struct{}),
		unknownTopicQueue:   make(chan *queueStruct, topicQueueSize),
		queueSize:           queueSize,
		topicQueueSize:      topicQueueSize,
		concurrentCalls:     concurrentCalls,
	}
	go topicManager.handleCalls()
	for topic, handler := range topicHandlers {
		topicManager.AddTopic(topic, handler)
	}
	return topicManager
}

func (topicManager *TopicManager) handleCalls() {
	for {
		queueStruct := <-topicManager.queue
		if queueStruct == nil {
			return
		}
		if queue := topicManager.topicQueues[queueStruct.topic]; queue != nil {
			queue <- queueStruct
		} else if topicManager.unknownTopicQueue != nil {
			topicManager.unknownTopicQueue <- queueStruct
		} else {
			queueStruct.responseAnyChannel <- nil
			queueStruct.responseErrorChannel <- errors.New("no handler for topic")
		}
	}
}

func (topicManager *TopicManager) handleTopic(queue chan *queueStruct, handler TopicHandler, stopChannel chan struct{}) {
	for queueStruct := range queue {
		select {
		case <-stopChannel:
			return
		default:
		}

		if topicManager.concurrentCalls {
			go func() {
				response, err := handler(queueStruct.args...)
				queueStruct.responseAnyChannel <- response
				queueStruct.responseErrorChannel <- err
			}()
		} else {
			response, err := handler(queueStruct.args...)
			queueStruct.responseAnyChannel <- response
			queueStruct.responseErrorChannel <- err
		}
	}
}

func (topicManager *TopicManager) HandleTopic(topic string, args ...any) (any, error) {
	response := make(chan any)
	err := make(chan error)

	topicManager.queue <- &queueStruct{
		topic:                topic,
		args:                 args,
		responseAnyChannel:   response,
		responseErrorChannel: err,
	}

	return <-response, <-err
}

func (topicManager *TopicManager) Close() error {

}

func (messageHandler *TopicManager) AddTopic(topic string, handler TopicHandler) error {
	messageHandler.mutex.Lock()
	defer messageHandler.mutex.Unlock()
	if messageHandler.isCLosed {
		return errors.New("topic manager is closed")
	}
	if messageHandler.topicQueues[topic] != nil {
		return errors.New("topic already exists")
	}
	queue := make(chan *queueStruct, messageHandler.topicQueueSize)
	stopChannel := make(chan struct{})
	messageHandler.topicQueues[topic] = queue
	messageHandler.topicHandlers[topic] = handler
	messageHandler.topicStopChannels[topic] = stopChannel
	go messageHandler.handleTopic(queue, handler, stopChannel)
	return nil
}

func (messageHandler *TopicManager) RemoveTopic(topic string) error {
	messageHandler.mutex.Lock()
	defer messageHandler.mutex.Unlock()
	if messageHandler.isCLosed {
		return errors.New("topic manager is closed")
	}
	if messageHandler.topicQueues[topic] == nil {
		return errors.New("topic does not exist")
	}
	close(messageHandler.topicQueues[topic])
	close(messageHandler.topicStopChannels[topic])

	delete(messageHandler.topicQueues, topic)
	delete(messageHandler.topicHandlers, topic)
	delete(messageHandler.topicStopChannels, topic)
	return nil
}

func (messageHandler *TopicManager) GetHandler(topic string) TopicHandler {
	return nil
}

func (messageHandler *TopicManager) GetUnknownHandler() TopicHandler {
	return nil
}

func (messageHandler *TopicManager) SetUnknownHandler() error {
	return nil
}

func (messageHandler *TopicManager) GetTopics() []string {
	return nil
}
