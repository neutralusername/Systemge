package Tools

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Config"
)

type TopicHandler func(...any) (any, error)
type TopicHandlers map[string]TopicHandler

type TopicManager struct {
	config *Config.TopicManager

	topicHandlers       TopicHandlers
	unknownTopicHandler TopicHandler

	isClosed bool
	mutex    sync.Mutex

	queue             chan *queueStruct
	topicQueues       map[string]chan *queueStruct
	unknownTopicQueue chan *queueStruct
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

func NewTopicManager(config *Config.TopicManager, topicHandlers TopicHandlers, unknownTopicHandler TopicHandler) *TopicManager {
	if topicHandlers == nil {
		topicHandlers = make(TopicHandlers)
	}
	topicManager := &TopicManager{
		config:              config,
		topicHandlers:       topicHandlers,
		unknownTopicHandler: unknownTopicHandler,
		queue:               make(chan *queueStruct, config.QueueSize),
		topicQueues:         make(map[string]chan *queueStruct),
	}
	go topicManager.handleCalls()
	for topic, handler := range topicHandlers {
		queue := make(chan *queueStruct, config.TopicQueueSize)
		topicManager.topicQueues[topic] = queue
		topicManager.topicHandlers[topic] = handler
		go topicManager.handleTopic(queue, handler)
	}
	if unknownTopicHandler != nil {
		topicManager.unknownTopicQueue = make(chan *queueStruct, config.TopicQueueSize)
		go topicManager.handleTopic(topicManager.unknownTopicQueue, unknownTopicHandler)
	}
	return topicManager
}

// can not be called after Close or will cause panic.
func (topicManager *TopicManager) HandleTopic(topic string, args ...any) (any, error) {
	queueStruct := &queueStruct{
		topic:                topic,
		args:                 args,
		responseAnyChannel:   make(chan any),
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
	return <-queueStruct.responseAnyChannel, <-queueStruct.responseErrorChannel
}

func (topicManager *TopicManager) handleCalls() {
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
		if topicManager.config.TopicQueueBlocking {
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

func (topicManager *TopicManager) handleTopic(queue chan *queueStruct, handler TopicHandler) {
	for queueStruct := range queue {
		if topicManager.config.ConcurrentCalls {
			go topicManager.handleCall(queueStruct, handler)
		} else {
			topicManager.handleCall(queueStruct, handler)
		}
	}
}
func (topicManager *TopicManager) handleCall(queueStruct *queueStruct, handler TopicHandler) {
	var deadline <-chan time.Time
	if topicManager.config.DeadlineMs > 0 {
		deadline = time.After(time.Duration(topicManager.config.DeadlineMs) * time.Millisecond)
	}
	var myCase chan bool = make(chan bool)
	go func() {
		response, err := handler(queueStruct.args...)
		queueStruct.responseAnyChannel <- response
		queueStruct.responseErrorChannel <- err
		close(myCase)
	}()
	select {
	case <-deadline:
		queueStruct.responseAnyChannel <- nil
		queueStruct.responseErrorChannel <- errors.New("deadline exceeded")
	case <-myCase:
	}
}

func (topicManager *TopicManager) Close() error {
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

func (topicManager *TopicManager) IsClosed() bool {
	topicManager.mutex.Lock()
	defer topicManager.mutex.Unlock()
	return topicManager.isClosed
}
