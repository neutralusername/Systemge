package WebsocketConnection

import (
	"sync"

	"github.com/gorilla/websocket"
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
}
