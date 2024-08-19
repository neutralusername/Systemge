package SystemgeConnection

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tcp"
)

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
			if Tcp.IsConnectionClosed(err) {
				connection.Close()
				return nil, Error.New("Connection closed", err)
			}
			return nil, err
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
		if Tcp.IsConnectionClosed(err) {
			connection.Close()
			return Error.New("Connection closed", err)
		}
		return err
	}
	connection.bytesSent.Add(bytesSent)
	return nil
}
