package TopicManager

import (
	"errors"
)

type TopicHandler func(...any) (any, error)
type TopicHandlers map[string]TopicHandler

type TopicManager struct {
	topicHandlers       TopicHandlers
	unknownTopicHandler TopicHandler

	isStarted bool

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
		unknownTopicQueue:   make(chan *queueStruct, topicQueueSize),
		queueSize:           queueSize,
		topicQueueSize:      topicQueueSize,
		concurrentCalls:     concurrentCalls,
	}
	go topicManager.handleCalls()
	for topic := range topicHandlers {
		topicManager.topicQueues[topic] = make(chan *queueStruct, topicQueueSize)
		go topicManager.handleTopic(topic)
	}
	return topicManager
}

func (topicManager *TopicManager) handleTopic(topic string) {
	queue := topicManager.topicQueues[topic]
	handler := topicManager.topicHandlers[topic]
	for {
		queueStruct := <-queue
		if queueStruct == nil {
			// change to cancel by closing channel
			return
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

// Caller is responsible for preventing concurrent calls to Start and Stop
func (topicManager *TopicManager) Start() error {

}

// Caller is responsible for preventing concurrent calls to Start and Stop
func (topicMananger *TopicManager) Stop() error {

}

func (messageHandler *TopicManager) AddTopic() error {
	return nil
}

func (messageHandler *TopicManager) RemoveTopic() error {
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
