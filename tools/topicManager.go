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

type TopicHandler[P any] func(P)
type TopicHandlers[P any] map[string]TopicHandler[P]

type TopicManager[P any] struct {
	config *configs.TopicManager

	topicHandlers       TopicHandlers[P]
	unknownTopicHandler TopicHandler[P]

	isClosed bool
	mutex    sync.Mutex

	queue             chan *queueStruct[P]
	topicQueues       map[string]chan *queueStruct[P]
	unknownTopicQueue chan *queueStruct[P]
}

type queueStruct[P any] struct {
	topic        string
	parameter    P
	errorChannel chan error
}

// modes: (l == large enough to never be full (depends on how many calls are made/how long they take to process))
// topicQueueSize: 0, queueSize: l concurrentCalls: false -> "sequential"
// topicQueueSize: l, queueSize: l concurrentCalls: false -> "topic exclusive"
// topicQueueSize: 0|l, queueSize: 0|l concurrentCalls: true -> "concurrent"

func NewTopicManager[P any](config *configs.TopicManager, topicHandlers TopicHandlers[P], unknownTopicHandler TopicHandler[P]) (*TopicManager[P], error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if topicHandlers == nil && unknownTopicHandler == nil {
		return nil, errors.New("no handlers provided")
	}

	if topicHandlers == nil {
		topicHandlers = make(TopicHandlers[P])
	}
	topicManager := &TopicManager[P]{
		config:              config,
		topicHandlers:       topicHandlers,
		unknownTopicHandler: unknownTopicHandler,
		queue:               make(chan *queueStruct[P], config.QueueSize),
		topicQueues:         make(map[string]chan *queueStruct[P]),
	}
	go topicManager.handleCalls()
	for topic, handler := range topicHandlers {
		queue := make(chan *queueStruct[P], config.TopicQueueSize)
		topicManager.topicQueues[topic] = queue
		topicManager.topicHandlers[topic] = handler
		go topicManager.handleTopic(queue, handler)
	}
	if unknownTopicHandler != nil {
		topicManager.unknownTopicQueue = make(chan *queueStruct[P], config.TopicQueueSize)
		go topicManager.handleTopic(topicManager.unknownTopicQueue, unknownTopicHandler)
	}
	return topicManager, nil
}

// can not be called after Close or will cause panic.
func (topicManager *TopicManager[P]) Handle(topic string, parameter P) error {
	queueStruct := &queueStruct[P]{
		topic:        topic,
		parameter:    parameter,
		errorChannel: make(chan error, 1),
	}

	if topicManager.config.QueueBlocking {
		topicManager.queue <- queueStruct
	} else {
		select {
		case topicManager.queue <- queueStruct:
		default:
			return errors.New("queue full")
		}
	}
	return <-queueStruct.errorChannel
}

func (topicManager *TopicManager[P]) handleCalls() {
	for queueStruct := range topicManager.queue {
		queue := topicManager.topicQueues[queueStruct.topic]
		if queue == nil {
			if topicManager.unknownTopicQueue != nil {
				queue = topicManager.unknownTopicQueue
			} else {
				queueStruct.errorChannel <- errors.New("no handler for topic")
				close(queueStruct.errorChannel)
				continue
			}
		}
		if topicManager.config.TopicQueueBlocking {
			queue <- queueStruct
		} else {
			select {
			case queue <- queueStruct:
			default:
				queueStruct.errorChannel <- errors.New("topic queue full")
				close(queueStruct.errorChannel)
			}
		}
	}
}

func (topicManager *TopicManager[P]) handleTopic(queue chan *queueStruct[P], handler TopicHandler[P]) {
	for queueStruct := range queue {
		if topicManager.config.ConcurrentCalls {
			go topicManager.handleCall(queueStruct, handler)
		} else {
			topicManager.handleCall(queueStruct, handler)
		}
	}
}
func (topicManager *TopicManager[P]) handleCall(queueStruct *queueStruct[P], handler TopicHandler[P]) {
	defer close(queueStruct.errorChannel)

	if topicManager.config.TimeoutNs == 0 {
		handler(queueStruct.parameter)
	}

	var callback chan struct{} = make(chan struct{})
	go func() {
		handler(queueStruct.parameter)
		close(callback)
	}()

	select {
	case <-time.After(time.Duration(topicManager.config.TimeoutNs) * time.Nanosecond):
		queueStruct.errorChannel <- errors.New("timeout")
	case <-callback:
	}
}

func (topicManager *TopicManager[P]) Close() error {
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

func (topicManager *TopicManager[P]) IsClosed() bool {
	topicManager.mutex.Lock()
	defer topicManager.mutex.Unlock()

	return topicManager.isClosed
}
