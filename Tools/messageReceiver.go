package Tools

import (
	"net"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tcp"
)

type MessageReceiver struct {
	buffer                   []byte
	incomingMessageByteLimit uint64
	tcpReceiveTimeoutMs      uint64
	bufferSize               uint32
	netConn                  net.Conn

	// metrics
	bytesReceived atomic.Uint64
}

func (buffer *MessageReceiver) GetBytesReceived() uint64 {
	return buffer.bytesReceived.Swap(0)
}

func (buffer *MessageReceiver) CheckBytesReceived() uint64 {
	return buffer.bytesReceived.Load()
}

func NewMessageReceiver(netConn net.Conn, incomingMessageByteLimit uint64, tcpReceiveTimeoutMs uint64, bufferSize uint32) *MessageReceiver {
	if bufferSize == 0 {
		bufferSize = 1024 * 4
	}
	return &MessageReceiver{
		buffer:                   []byte{},
		incomingMessageByteLimit: incomingMessageByteLimit,
		netConn:                  netConn,
		tcpReceiveTimeoutMs:      tcpReceiveTimeoutMs,
		bufferSize:               bufferSize,
	}
}

func (buffer *MessageReceiver) ReceiveNextMessage() ([]byte, error) {
	completedMsgBytes := []byte{}
	for {
		if buffer.incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > buffer.incomingMessageByteLimit {
			return nil, Error.New("Incoming message byte limit exceeded", nil)
		}
		for i, b := range buffer.buffer {
			if b == Tcp.HEARTBEAT {
				continue
			}
			if b == Tcp.ENDOFMESSAGE {
				buffer.buffer = buffer.buffer[i+1:]
				if buffer.incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > buffer.incomingMessageByteLimit {
					// i am considering removing this error case and just returning the message instead, even though the limit is exceeded, but only by less than the buffer size
					return nil, Error.New("Incoming message byte limit exceeded", nil)
				}
				return completedMsgBytes, nil
			}
			completedMsgBytes = append(completedMsgBytes, b)
		}
		receivedMessageBytes, _, err := Tcp.Receive(buffer.netConn, buffer.tcpReceiveTimeoutMs, buffer.bufferSize)
		if err != nil {
			return nil, err
		}
		buffer.buffer = receivedMessageBytes
		buffer.bytesReceived.Add(uint64(len(receivedMessageBytes)))
	}
}
