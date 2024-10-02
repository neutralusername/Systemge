package WebsocketClient

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

type WebsocketClient struct {
	id            string
	config        *Config.WebsocketConnection
	websocketConn *websocket.Conn

	closed       bool
	closedMutex  sync.Mutex
	closeChannel chan bool
	waitGroup    sync.WaitGroup

	sendMutex sync.Mutex

	messageHandlingLoopStopChannel chan<- bool
	messageMutex                   sync.Mutex
	messageChannel                 chan *Message.Message
	messageChannelSemaphore        *Tools.Semaphore

	byteRateLimiter    *Tools.TokenBucketRateLimiter
	messageRateLimiter *Tools.TokenBucketRateLimiter

	eventHandler Event.Handler

	// metrics

	bytesSent     atomic.Uint64
	bytesReceived atomic.Uint64

	messagesSent     atomic.Uint64
	messagesReceived atomic.Uint64

	invalidMessagesReceived  atomic.Uint64
	rejectedMessagesReceived atomic.Uint64
}

func New(id string, config *Config.WebsocketConnection, websocketConn *websocket.Conn, eventHandler Event.Handler) (*WebsocketClient, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if websocketConn == nil {
		return nil, errors.New("netConn is nil")
	}

	connection := &WebsocketClient{
		id:                      id,
		config:                  config,
		websocketConn:           websocketConn,
		closeChannel:            make(chan bool),
		messageChannel:          make(chan *Message.Message, config.MessageChannelCapacity+1), // +1 so that the receive loop is never blocking while adding a message to the processing channel
		messageChannelSemaphore: Tools.NewSemaphore(config.MessageChannelCapacity+1, config.MessageChannelCapacity+1),
		eventHandler:            eventHandler,
	}
	if config.RateLimiterBytes != nil {
		connection.byteRateLimiter = Tools.NewTokenBucketRateLimiter(config.RateLimiterBytes)
	}
	if config.RateLimiterMessages != nil {
		connection.messageRateLimiter = Tools.NewTokenBucketRateLimiter(config.RateLimiterMessages)
	}

	connection.waitGroup.Add(1)
	go connection.receptionRoutine()
	return connection, nil
}
