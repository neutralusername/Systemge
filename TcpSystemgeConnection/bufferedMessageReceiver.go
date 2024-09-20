package TcpSystemgeConnection

import (
	"net"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tcp"
)

type BufferedMessageReceiver struct {
	buffer                   []byte
	incomingMessageByteLimit uint64
	tcpReceiveTimeoutMs      uint64
	bufferSize               uint32
	netConn                  net.Conn

	// metrics
	bytesReceived atomic.Uint64
}

func (buffer *BufferedMessageReceiver) GetBytesReceived() uint64 {
	return buffer.bytesReceived.Swap(0)
}

func (buffer *BufferedMessageReceiver) CheckBytesReceived() uint64 {
	return buffer.bytesReceived.Load()
}

func NewBufferedMessageReceiver(netConn net.Conn, incomingMessageByteLimit uint64, tcpReceiveTimeoutMs uint64, bufferSize uint32) *BufferedMessageReceiver {
	if bufferSize == 0 {
		bufferSize = 1024 * 4
	}
	return &BufferedMessageReceiver{
		buffer:                   []byte{},
		incomingMessageByteLimit: incomingMessageByteLimit,
		netConn:                  netConn,
		tcpReceiveTimeoutMs:      tcpReceiveTimeoutMs,
		bufferSize:               bufferSize,
	}
}

func (messageReceiver *BufferedMessageReceiver) ReceiveNextMessage() ([]byte, error) {
	completedMsgBytes := []byte{}
	for {
		if messageReceiver.incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > messageReceiver.incomingMessageByteLimit {
			return nil, Event.New("Incoming message byte limit exceeded", nil)
		}
		for i, b := range messageReceiver.buffer {
			if b == Tcp.HEARTBEAT {
				continue
			}
			if b == Tcp.ENDOFMESSAGE {
				messageReceiver.buffer = messageReceiver.buffer[i+1:]
				if messageReceiver.incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > messageReceiver.incomingMessageByteLimit {
					// i am considering removing this error case and just returning the message instead, even though the limit is exceeded, but only by less than the buffer size
					return nil, Event.New("Incoming message byte limit exceeded", nil)
				}
				return completedMsgBytes, nil
			}
			completedMsgBytes = append(completedMsgBytes, b)
		}
		receivedMessageBytes, _, err := Tcp.Receive(messageReceiver.netConn, messageReceiver.tcpReceiveTimeoutMs, messageReceiver.bufferSize)
		if err != nil {
			return nil, err
		}
		messageReceiver.buffer = receivedMessageBytes
		messageReceiver.bytesReceived.Add(uint64(len(receivedMessageBytes)))
	}
}
