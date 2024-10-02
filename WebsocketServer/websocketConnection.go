package WebsocketServer

import (
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

type WebsocketConnection struct {
	id            string
	websocketConn *websocket.Conn

	sendMutex   sync.Mutex
	stopChannel chan bool
	waitGroup   sync.WaitGroup

	closeMutex sync.Mutex
	isClosed   bool

	byteRateLimiter    *Tools.TokenBucketRateLimiter
	messageRateLimiter *Tools.TokenBucketRateLimiter
}

func (server *WebsocketServer) NewWebsocketConnection(websocketConn *websocket.Conn) *WebsocketConnection {
	connection := &WebsocketConnection{
		websocketConn: websocketConn,
		stopChannel:   make(chan bool),
	}
	websocketConn.SetReadLimit(int64(server.config.IncomingMessageByteLimit))
	return connection
}

// Disconnects the websocketConnection and blocks until the websocketConnection onDisconnectHandler has finished.
func (websocketConnection *WebsocketConnection) Close() error {
	websocketConnection.closeMutex.Lock()
	defer websocketConnection.closeMutex.Unlock()
	if websocketConnection.isClosed {
		return errors.New("websocketConnection is already closed")
	}
	websocketConnection.isClosed = true
	websocketConnection.websocketConn.Close()
	if websocketConnection.byteRateLimiter != nil {
		websocketConnection.byteRateLimiter.Close()
	}
	if websocketConnection.messageRateLimiter != nil {
		websocketConnection.messageRateLimiter.Close()
	}
	close(websocketConnection.stopChannel)
	return nil
}

// Returns the ip of the websocketConnection.
func (websocketConnection *WebsocketConnection) GetAddress() string {
	return websocketConnection.websocketConn.RemoteAddr().String()
}

func (server *WebsocketServer) Write(websocketConnection *WebsocketConnection, messageBytes []byte) error {
	return server.write(websocketConnection, messageBytes, Event.SendRuntime)
}
func (server *WebsocketServer) write(websocketConnection *WebsocketConnection, messageBytes []byte, circumstance string) error {
	websocketConnection.sendMutex.Lock()
	defer websocketConnection.sendMutex.Unlock()

	if server.eventHandler != nil {
		if event := server.onEvent(Event.NewInfo(
			Event.WritingMessage,
			"sending websocketConnection message",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: circumstance,
				Event.Identity:     websocketConnection.GetId(),
				Event.Address:      websocketConnection.GetAddress(),
				Event.Bytes:        string(messageBytes),
			},
		)); !event.IsInfo() {
			return event.GetError()
		}
	}

	err := websocketConnection.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		if server.eventHandler != nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.WriteMessageFailed,
				err.Error(),
				Event.Context{
					Event.Circumstance: circumstance,
					Event.Identity:     websocketConnection.GetId(),
					Event.Address:      websocketConnection.GetAddress(),
					Event.Bytes:        string(messageBytes),
				}),
			)
		}
		return err
	}
	server.websocketConnectionMessagesSent.Add(1)
	server.websocketConnectionMessagesBytesSent.Add(uint64(len(messageBytes)))

	if server.eventHandler != nil {
		server.onEvent(Event.NewInfoNoOption(
			Event.WroteMessage,
			"sent websocketConnection message",
			Event.Context{
				Event.Circumstance: circumstance,
				Event.Identity:     websocketConnection.GetId(),
				Event.Address:      websocketConnection.GetAddress(),
				Event.Bytes:        string(messageBytes),
			}),
		)
	}
	return nil
}

// Returns the id of the websocketConnection.
func (websocketConnection *WebsocketConnection) GetId() string {
	return websocketConnection.id
}
