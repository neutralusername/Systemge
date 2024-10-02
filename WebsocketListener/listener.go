package WebsocketListener

import (
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Tools"
)

type WebsocketListener struct {
	name string

	isClosed    bool
	closedMutex sync.Mutex

	config        *Config.TcpSystemgeListener
	ipRateLimiter *Tools.IpRateLimiter

	httpServer        *HTTPServer.HTTPServer
	connectionChannel chan *websocket.Conn

	blacklist *Tools.AccessControlList
	whitelist *Tools.AccessControlList

	eventHandler Event.Handler

	// metrics

	accepted atomic.Uint32
	failed   atomic.Uint32
	rejected atomic.Uint32
}
