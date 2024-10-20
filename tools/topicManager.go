package tools

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/systemge/configs"
)

/*
type AsyncObjectHandler[O any, C any] func(object O, caller C) error
type SyncObjectHandler[O any, C any] func(object O, caller C, syncToken string) (O, error)
*/

type TopicHandler[P any, R any] func(P) (R, error)
type TopicHandlers[P any, R any] map[string]TopicHandler[P, R]

type TopicManager[P any, R any] struct {
	config *configs.TopicManager

	topicHandlers       TopicHandlers[P, R]
	unknownTopicHandler TopicHandler[P, R]

	isClosed bool
	mutex    sync.Mutex

	queue             chan *queueStruct[P, R]
	topicQueues       map[string]chan *queueStruct[P, R]
	unknownTopicQueue chan *queueStruct[P, R]
}

type queueStruct[P any, R any] struct {
	topic              string
	parameter          P
	resultChannel      chan R
	resultErrorChannel chan error
}

// modes: (l == large enough to never be full (depends on how many calls are made/how long they take to process))
// topicQueueSize: 0, queueSize: l concurrentCalls: false -> "sequential"
// topicQueueSize: l, queueSize: l concurrentCalls: false -> "topic exclusive"
// topicQueueSize: 0|l, queueSize: 0|l concurrentCalls: true -> "concurrent"

func NewTopicManager[P any, R any](config *configs.TopicManager, topicHandlers TopicHandlers[P, R], unknownTopicHandler TopicHandler[P, R]) *TopicManager[P, R] {
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
func (topicManager *TopicManager[P, R]) Handle(topic string, parameter P) (R, error) {
	queueStruct := &queueStruct[P, R]{
		topic:              topic,
		parameter:          parameter,
		resultChannel:      make(chan R),
		resultErrorChannel: make(chan error),
	}

	if topicManager.config.QueueBlocking {
		topicManager.queue <- queueStruct
	} else {
		select {
		case topicManager.queue <- queueStruct:
		default:
			close(queueStruct.resultChannel)
			return <-queueStruct.resultChannel, errors.New("queue full")
		}
	}
	return <-queueStruct.resultChannel, <-queueStruct.resultErrorChannel
}

func (topicManager *TopicManager[P, R]) handleCalls() {
	for queueStruct := range topicManager.queue {
		queue := topicManager.topicQueues[queueStruct.topic]
		if queue == nil {
			if topicManager.unknownTopicQueue != nil {
				queue = topicManager.unknownTopicQueue
			} else {
				close(queueStruct.resultChannel)
				queueStruct.resultErrorChannel <- errors.New("no handler for topic")
				continue
			}
		}
		if topicManager.config.TopicQueueBlocking {
			queue <- queueStruct
		} else {
			select {
			case queue <- queueStruct:
			default:
				close(queueStruct.resultChannel)
				queueStruct.resultErrorChannel <- errors.New("topic queue full")
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
	if topicManager.config.TimeoutNs > 0 {
		var callback chan struct{} = make(chan struct{})
		go func() {
			result, err := handler(queueStruct.parameter)
			queueStruct.resultChannel <- result
			queueStruct.resultErrorChannel <- err
			close(callback)
		}()
		select {
		case <-time.After(time.Duration(topicManager.config.TimeoutNs) * time.Nanosecond):
			close(queueStruct.resultChannel)
			queueStruct.resultErrorChannel <- errors.New("deadline exceeded")
		case <-callback:
		}
	} else {
		result, err := handler(queueStruct.parameter)
		queueStruct.resultChannel <- result
		queueStruct.resultErrorChannel <- err
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
