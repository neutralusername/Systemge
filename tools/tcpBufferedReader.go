package tools

import (
	"errors"
	"net"

	"github.com/neutralusername/systemge/configs"
)

const ENDOFMESSAGE = '\x04'
const HEARTBEAT = '\x05'

type TcpBufferedReader struct {
	config  *configs.TcpBufferedReader
	buffer  []byte
	netConn net.Conn
}

func NewTcpBufferedReader(netConn net.Conn, config *configs.TcpBufferedReader) *TcpBufferedReader {
	if config.BufferBytes == 0 {
		config.BufferBytes = 1024 * 4
	}
	return &TcpBufferedReader{
		config:  config,
		buffer:  []byte{},
		netConn: netConn,
	}
}

func (messageReceiver *TcpBufferedReader) Read() ([]byte, int, error) {
	completedMsgBytes := []byte{}
	newBytesRead := 0
	for {
		if messageReceiver.config.IncomingDataByteLimit > 0 && uint64(len(completedMsgBytes)) > messageReceiver.config.IncomingDataByteLimit {
			return nil, newBytesRead, errors.New("incoming message byte limit exceeded")
		}
		for i, b := range messageReceiver.buffer {
			if b == HEARTBEAT {
				continue
			}
			if b == ENDOFMESSAGE {
				messageReceiver.buffer = messageReceiver.buffer[i+1:]
				if messageReceiver.config.IncomingDataByteLimit > 0 && uint64(len(completedMsgBytes)) > messageReceiver.config.IncomingDataByteLimit {
					// i am considering removing this error case and just returning the message instead, even though the limit is exceeded, but only by less than the buffer size
					return nil, newBytesRead, errors.New("incoming message byte limit exceeded")
				}
				return completedMsgBytes, newBytesRead, nil
			}
			completedMsgBytes = append(completedMsgBytes, b)
		}

		buffer := make([]byte, messageReceiver.config.BufferBytes)
		newBytesReceived, err := messageReceiver.netConn.Read(buffer)
		if err != nil {
			return nil, newBytesRead, err
		}
		newBytesRead += newBytesReceived
		messageReceiver.buffer = buffer[:newBytesReceived]
	}
}
