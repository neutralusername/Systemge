package TopicManager

import (
	"errors"
	"sync"
)

type TopicHandler func(...any) (any, error)
type TopicHandlers map[string]TopicHandler

type Manager struct {
	topicHandlers       TopicHandlers
	unknownTopicHandler TopicHandler

	isClosed bool
	mutex    sync.Mutex

	queue             chan *queueStruct
	topicQueues       map[string]chan *queueStruct
	unknownTopicQueue chan *queueStruct

	queueSize          uint32
	topicQueueSize     uint32
	concurrentCalls    bool
	queueBlocking      bool
	topicQueueBlocking bool
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

func NewTopicManager(topicHandlers TopicHandlers, unknownTopicHandler TopicHandler, topicQueueSize uint32, queueSize uint32, concurrentCalls bool, queueBlocking bool, topicQueueBlocking bool) *Manager {
	if topicHandlers == nil {
		topicHandlers = make(TopicHandlers)
	}
	topicManager := &Manager{
		topicHandlers:       topicHandlers,
		unknownTopicHandler: unknownTopicHandler,
		queue:               make(chan *queueStruct, queueSize),
		topicQueues:         make(map[string]chan *queueStruct),
		queueSize:           queueSize,
		topicQueueSize:      topicQueueSize,
		concurrentCalls:     concurrentCalls,
		queueBlocking:       queueBlocking,
		topicQueueBlocking:  topicQueueBlocking,
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

// can not be called after Close or will cause panic.
func (topicManager *Manager) HandleTopic(topic string, args ...any) (any, error) {
	queueStruct := &queueStruct{
		topic:                topic,
		args:                 args,
		responseAnyChannel:   make(chan any),
		responseErrorChannel: make(chan error),
	}

	if topicManager.queueBlocking {
		topicManager.queue <- queueStruct
	} else {
		select {
		case topicManager.queue <- queueStruct:
		default:
			return nil, errors.New("queue full")
		}
	}
	return <-queueStruct.responseAnyChannel, <-queueStruct.responseErrorChannel
}

func (topicManager *Manager) handleCalls() {
	for queueStruct := range topicManager.queue {
		queue := topicManager.topicQueues[queueStruct.topic]
		if queue == nil {
			if topicManager.unknownTopicQueue != nil {
				queue = topicManager.unknownTopicQueue
			} else {
				queueStruct.responseAnyChannel <- nil
				queueStruct.responseErrorChannel <- errors.New("no handler for topic")
				continue
			}
		}
		if topicManager.topicQueueBlocking {
			queue <- queueStruct
		} else {
			select {
			case queue <- queueStruct:
			default:
				queueStruct.responseAnyChannel <- nil
				queueStruct.responseErrorChannel <- errors.New("topic queue full")
			}
		}
	}
}

func (topicManager *Manager) handleTopic(queue chan *queueStruct, handler TopicHandler) {
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

func (topicManager *Manager) Close() error {
	topicManager.mutex.Lock()
	defer topicManager.mutex.Unlock()

	if topicManager.isClosed {
		return errors.New("topic manager already closed")
	}

	topicManager.isClosed = true
	close(topicManager.queue)
	for _, queue := range topicManager.topicQueues {
		close(queue)
	}
	close(topicManager.unknownTopicQueue)

	return nil
}

func (topicManager *Manager) IsClosed() bool {
	topicManager.mutex.Lock()
	defer topicManager.mutex.Unlock()
	return topicManager.isClosed
}
