package Tools

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Config"
)

type TopicHandler[P any, R any] func(P) (R, error)
type TopicHandlers[P any, R any] map[string]TopicHandler[P, R]

type TopicManager[P any, R any] struct {
	config *Config.TopicManager

	topicHandlers       TopicHandlers[P, R]
	unknownTopicHandler TopicHandler[P, R]

	isClosed bool
	mutex    sync.Mutex

	queue             chan *queueStruct[P, R]
	topicQueues       map[string]chan *queueStruct[P, R]
	unknownTopicQueue chan *queueStruct[P, R]
}

type queueStruct[P any, R any] struct {
	topic                string
	parameter            P
	responseChanneö      chan R
	responseErrorChannel chan error
}

// modes: (l == large enough to never be full (depends on how many calls are made/how long they take to process))
// topicQueueSize: 0, queueSize: l concurrentCalls: false -> "sequential"
// topicQueueSize: l, queueSize: l concurrentCalls: false -> "topic exclusive"
// topicQueueSize: 0|l, queueSize: 0|l concurrentCalls: true -> "concurrent"

func NewTopicManager[P any, R any](config *Config.TopicManager, topicHandlers TopicHandlers[P, R], unknownTopicHandler TopicHandler[P, R]) *TopicManager[P, R] {
	if topicHandlers == nil {
		topicHandlers = make(TopicHandlers[P, R])
	}
	topicManager := &TopicManager[P, R]{
		config:              config,
		topicHandlers:       topicHandlers,
		unknownTopicHandler: unknownTopicHandler,
		queue:               make(chan *queueStruct[P, R], config.QueueSize),
		topicQueues:         make(map[string]chan *queueStruct[P, R]),
	}
	go topicManager.handleCalls()
	for topic, handler := range topicHandlers {
		queue := make(chan *queueStruct[P, R], config.TopicQueueSize)
		topicManager.topicQueues[topic] = queue
		topicManager.topicHandlers[topic] = handler
		go topicManager.handleTopic(queue, handler)
	}
	if unknownTopicHandler != nil {
		topicManager.unknownTopicQueue = make(chan *queueStruct[P, R], config.TopicQueueSize)
		go topicManager.handleTopic(topicManager.unknownTopicQueue, unknownTopicHandler)
	}
	return topicManager
}

// can not be called after Close or will cause panic.
func (topicManager *TopicManager[P, R]) Handle(topic string, parameter P) (any, error) {
	queueStruct := &queueStruct[P, R]{
		topic:                topic,
		parameter:            parameter,
		responseChanneö:      make(chan R),
		responseErrorChannel: make(chan error),
	}

	if topicManager.config.QueueBlocking {
		topicManager.queue <- queueStruct
	} else {
		select {
		case topicManager.queue <- queueStruct:
		default:
			return nil, errors.New("queue full")
		}
	}
	return <-queueStruct.responseChanneö, <-queueStruct.responseErrorChannel
}

func (topicManager *TopicManager[P, R]) handleCalls() {
	for queueStruct := range topicManager.queue {
		queue := topicManager.topicQueues[queueStruct.topic]
		if queue == nil {
			if topicManager.unknownTopicQueue != nil {
				queue = topicManager.unknownTopicQueue
			} else {
				close(queueStruct.responseChanneö)
				queueStruct.responseErrorChannel <- errors.New("no handler for topic")
				continue
			}
		}
		if topicManager.config.TopicQueueBlocking {
			queue <- queueStruct
		} else {
			select {
			case queue <- queueStruct:
			default:
				close(queueStruct.responseChanneö)
				queueStruct.responseErrorChannel <- errors.New("topic queue full")
			}
		}
	}
}

func (topicManager *TopicManager[P, R]) handleTopic(queue chan *queueStruct[P, R], handler TopicHandler[P, R]) {
	for queueStruct := range queue {
		if topicManager.config.ConcurrentCalls {
			go topicManager.handleCall(queueStruct, handler)
		} else {
			topicManager.handleCall(queueStruct, handler)
		}
	}
}
func (topicManager *TopicManager[P, R]) handleCall(queueStruct *queueStruct[P, R], handler TopicHandler[P, R]) {
	if topicManager.config.TimeoutMs > 0 {
		var callback chan struct{} = make(chan struct{})
		go func() {
			response, err := handler(queueStruct.parameter)
			queueStruct.responseChanneö <- response
			queueStruct.responseErrorChannel <- err
			close(callback)
		}()
		select {
		case <-time.After(time.Duration(topicManager.config.TimeoutMs) * time.Millisecond):
			close(queueStruct.responseChanneö)
			queueStruct.responseErrorChannel <- errors.New("deadline exceeded")
		case <-callback:
		}
	} else {
		response, err := handler(queueStruct.parameter)
		queueStruct.responseChanneö <- response
		queueStruct.responseErrorChannel <- err
	}
}

func (topicManager *TopicManager[P, R]) Close() error {
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

func (topicManager *TopicManager[P, R]) IsClosed() bool {
	topicManager.mutex.Lock()
	defer topicManager.mutex.Unlock()
	return topicManager.isClosed
}
