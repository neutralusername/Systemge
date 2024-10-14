package TcpConnection

import (
	"errors"
	"net"
	"sync/atomic"
	"time"
)

type BufferedMessageReader struct {
	buffer                   []byte
	incomingMessageByteLimit uint64
	tcpReceiveTimeoutNs      int64
	bufferSize               uint32
	netConn                  net.Conn

	// metrics
	bytesReceived atomic.Uint64
}

func (buffer *BufferedMessageReader) GetBytesReceived() uint64 {
	return buffer.bytesReceived.Swap(0)
}

func (buffer *BufferedMessageReader) CheckBytesReceived() uint64 {
	return buffer.bytesReceived.Load()
}

func NewBufferedMessageReader(netConn net.Conn, incomingMessageByteLimit uint64, tcpReceiveTimeoutNs int64, bufferSize uint32) *BufferedMessageReader {
	if bufferSize == 0 {
		bufferSize = 1024 * 4
	}
	return &BufferedMessageReader{
		buffer:                   []byte{},
		incomingMessageByteLimit: incomingMessageByteLimit,
		netConn:                  netConn,
		tcpReceiveTimeoutNs:      tcpReceiveTimeoutNs,
		bufferSize:               bufferSize,
	}
}

func (messageReceiver *BufferedMessageReader) ReadNextMessage() ([]byte, error) {
	completedMsgBytes := []byte{}
	for {
		if messageReceiver.incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > messageReceiver.incomingMessageByteLimit {
			return nil, errors.New("incoming message byte limit exceeded")
		}
		for i, b := range messageReceiver.buffer {
			if b == HEARTBEAT {
				continue
			}
			if b == ENDOFMESSAGE {
				messageReceiver.buffer = messageReceiver.buffer[i+1:]
				if messageReceiver.incomingMessageByteLimit > 0 && uint64(len(completedMsgBytes)) > messageReceiver.incomingMessageByteLimit {
					// i am considering removing this error case and just returning the message instead, even though the limit is exceeded, but only by less than the buffer size
					return nil, errors.New("incoming message byte limit exceeded")
				}
				return completedMsgBytes, nil
			}
			completedMsgBytes = append(completedMsgBytes, b)
		}

		messageReceiver.netConn.SetReadDeadline(time.Now().Add(time.Duration(messageReceiver.tcpReceiveTimeoutNs) * time.Nanosecond))
		buffer := make([]byte, messageReceiver.bufferSize)
		newBytesReceived, err := messageReceiver.netConn.Read(buffer)
		if err != nil {
			return nil, err
		}
		messageReceiver.buffer = buffer
		messageReceiver.bytesReceived.Add(uint64(newBytesReceived))
	}
}
