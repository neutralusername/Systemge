package WebsocketConnection

import (
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

type WebsocketConnection struct {
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
