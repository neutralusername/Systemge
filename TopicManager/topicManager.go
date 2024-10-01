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

	isCLosed   bool
	closeMutex sync.Mutex

	queue             chan *queueStruct
	topicQueues       map[string]chan *queueStruct
	unknownTopicQueue chan *queueStruct

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
		queueSize:           queueSize,
		topicQueueSize:      topicQueueSize,
		concurrentCalls:     concurrentCalls,
	}
	go topicManager.handleCalls()
	for topic, handler := range topicHandlers {
		queue := make(chan *queueStruct, topicManager.topicQueueSize)
		topicManager.topicQueues[topic] = queue
		topicManager.topicHandlers[topic] = handler
		go topicManager.handleTopic(queue, handler)
	}
	if unknownTopicHandler != nil {
		topicManager.unknownTopicQueue = make(chan *queueStruct, queueSize)
		go topicManager.handleTopic(topicManager.unknownTopicQueue, unknownTopicHandler)
	}
	return topicManager
}

func (topicManager *TopicManager) handleCalls() {
	for queueStruct := range topicManager.queue {
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

func (topicManager *TopicManager) handleTopic(queue chan *queueStruct, handler TopicHandler) {
	for queueStruct := range queue {
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

// can not be called after Close or will cause panic.
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
	topicManager.closeMutex.Lock()
	defer topicManager.closeMutex.Unlock()

	if topicManager.isCLosed {
		return errors.New("topic manager already closed")
	}

	topicManager.isCLosed = true
	close(topicManager.queue)
	for _, queue := range topicManager.topicQueues {
		close(queue)
	}
	close(topicManager.unknownTopicQueue)

	return nil
}
