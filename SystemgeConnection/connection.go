package SystemgeConnection

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeConnection struct {
	config     *Config.SystemgeConnection
	netConn    net.Conn
	randomizer *Tools.Randomizer

	sendMutex    sync.Mutex
	receiveMutex sync.Mutex

	rateLimiterBytes *Tools.TokenBucketRateLimiter
	rateLimiterMsgs  *Tools.TokenBucketRateLimiter
	tcpBuffer        []byte

	syncResponseChannels map[string]chan *Message.Message
	syncMutex            sync.Mutex

	stopChannel chan bool

	// metrics
	bytesSent     atomic.Uint64
	bytesReceived atomic.Uint64
}

func New(config *Config.SystemgeConnection, netConn net.Conn) *SystemgeConnection {
	connection := &SystemgeConnection{
		config:      config,
		netConn:     netConn,
		randomizer:  Tools.NewRandomizer(config.RandomizerSeed),
		stopChannel: make(chan bool),
	}
	if config.RateLimiterBytes != nil {
		connection.rateLimiterBytes = Tools.NewTokenBucketRateLimiter(config.RateLimiterBytes)
	}
	if config.RateLimiterMessages != nil {
		connection.rateLimiterMsgs = Tools.NewTokenBucketRateLimiter(config.RateLimiterMessages)
	}
	return connection
}

func (connection *SystemgeConnection) ReceiveMessage(bufferSize uint32, incomingMessageByteLimit uint64) ([]byte, error) {
	connection.receiveMutex.Lock()
	defer connection.receiveMutex.Unlock()

	completedMsgBytes := []byte{}
	for {
		if incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > incomingMessageByteLimit {
			return nil, Error.New("Incoming message byte limit exceeded", nil)
		}
		for i, b := range connection.tcpBuffer {
			if b == Tcp.ENDOFMESSAGE {
				completedMsgBytes = append(completedMsgBytes, connection.tcpBuffer[:i]...)
				connection.tcpBuffer = connection.tcpBuffer[i+1:]
				if incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > incomingMessageByteLimit {
					return nil, Error.New("Incoming message byte limit exceeded", nil)
				}
				return completedMsgBytes, nil
			}
		}
		completedMsgBytes = append(completedMsgBytes, connection.tcpBuffer...)
		receivedMessageBytes, _, err := Tcp.Receive(connection.netConn, connection.config.TcpReceiveTimeoutMs, bufferSize)
		if err != nil {
			return nil, Error.New("Failed to refill tcp buffer", err)
		}
		connection.tcpBuffer = receivedMessageBytes
		connection.bytesReceived.Add(uint64(len(receivedMessageBytes)))
	}
}

func (connection *SystemgeConnection) SendMessage(bytes []byte) error {
	connection.sendMutex.Lock()
	defer connection.sendMutex.Unlock()

	bytesSent, err := Tcp.Send(connection.netConn, bytes, connection.config.TcpSendTimeoutMs)
	if err != nil {
		return Error.New("Failed to send message", err)
	}
	connection.bytesSent.Add(bytesSent)
	return nil
}
