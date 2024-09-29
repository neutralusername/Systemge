package Tcp

import (
	"errors"
	"io"
	"net"
	"os"
	"syscall"
	"time"
)

const ENDOFMESSAGE = '\x04'
const HEARTBEAT = '\x05'

func Write(netConn net.Conn, bytes []byte, timeoutMs uint64) (uint64, error) {
	if netConn == nil {
		return 0, errors.New("net.Conn is nil")
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

func SendHeartbeat(netConn net.Conn, timeoutMs uint64) error {
	if netConn == nil {
		return errors.New("net.Conn is nil")
	}
	if timeoutMs > 0 {
		netConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	} else {
		netConn.SetWriteDeadline(time.Time{})
	}
	_, err := netConn.Write([]byte{HEARTBEAT})
	if err != nil {
		return err
	}
	return nil
}

func Read(netConn net.Conn, timeoutMs uint64, bufferSize uint32) ([]byte, int, error) {
	if netConn == nil {
		return nil, 0, errors.New("net.Conn is nil")
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

func IsConnectionClosed(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || errors.Is(err, os.ErrClosed) {
		return true
	}
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		switch netErr.Err {
		case syscall.EPIPE, syscall.ECONNRESET, syscall.ENOTCONN,
			syscall.ECONNABORTED, syscall.ENETDOWN,
			syscall.ENETUNREACH, syscall.EHOSTDOWN:
			return true
		}
	}
	if netErrInterface, ok := err.(net.Error); ok && !netErrInterface.Timeout() {
		return true
	}
	return false
}
