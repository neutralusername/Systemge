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
	name       string
	config     *Config.SystemgeConnection
	netConn    net.Conn
	randomizer *Tools.Randomizer

	sendMutex    sync.Mutex
	receiveMutex sync.Mutex

	tcpBuffer []byte

	syncResponseChannels map[string]chan *Message.Message
	syncMutex            sync.Mutex

	stopChannel chan bool

	// metrics
	bytesSent     atomic.Uint64
	bytesReceived atomic.Uint64

	asyncMessagesSent atomic.Uint32
	syncRequestsSent  atomic.Uint32

	syncSuccessResponsesReceived atomic.Uint32
	syncFailureResponsesReceived atomic.Uint32
	noSyncResponseReceived       atomic.Uint32
}

func New(config *Config.SystemgeConnection, netConn net.Conn, name string) *SystemgeConnection {
	connection := &SystemgeConnection{
		name:        name,
		config:      config,
		netConn:     netConn,
		randomizer:  Tools.NewRandomizer(config.RandomizerSeed),
		stopChannel: make(chan bool),
	}
	return connection
}

func (connection *SystemgeConnection) GetName() string {
	return connection.name
}

func (connection *SystemgeConnection) ReceiveMessage() ([]byte, error) {
	connection.receiveMutex.Lock()
	defer connection.receiveMutex.Unlock()

	completedMsgBytes := []byte{}
	for {
		if connection.config.IncomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > connection.config.IncomingMessageByteLimit {
			return nil, Error.New("Incoming message byte limit exceeded", nil)
		}
		for i, b := range connection.tcpBuffer {
			if b == Tcp.ENDOFMESSAGE {
				completedMsgBytes = append(completedMsgBytes, connection.tcpBuffer[:i]...)
				connection.tcpBuffer = connection.tcpBuffer[i+1:]
				if connection.config.IncomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > connection.config.IncomingMessageByteLimit {
					return nil, Error.New("Incoming message byte limit exceeded", nil)
				}
				return completedMsgBytes, nil
			}
		}
		completedMsgBytes = append(completedMsgBytes, connection.tcpBuffer...)
		receivedMessageBytes, _, err := Tcp.Receive(connection.netConn, connection.config.TcpReceiveTimeoutMs, connection.config.TcpBufferBytes)
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
