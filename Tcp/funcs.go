package Tcp

import (
	"net"
	"time"

	"github.com/neutralusername/Systemge/Error"
)

const ENDOFMESSAGE = '\x04'

func Send(netConn net.Conn, bytes []byte, timeoutMs uint64) (uint64, error) {
	if netConn == nil {
		return 0, Error.New("net.Conn is nil", nil)
	}
	if timeoutMs > 0 {
		netConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	} else {
		netConn.SetWriteDeadline(time.Time{})
	}
	bytesSend, err := netConn.Write(append(bytes, ENDOFMESSAGE))
	if err != nil {
		return 0, err
	}
	return uint64(bytesSend), nil
}

func Receive(netConn net.Conn, timeoutMs uint64, bufferSize uint32) ([]byte, int, error) {
	if netConn == nil {
		return nil, 0, Error.New("net.Conn is nil", nil)
	}
	buffer := make([]byte, bufferSize)
	if timeoutMs > 0 {
		netConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	} else {
		netConn.SetReadDeadline(time.Time{})
	}
	bytesRead, err := netConn.Read(buffer)
	if err != nil {
		return nil, bytesRead, err
	}
	return buffer[:bytesRead], bytesRead, nil
}
