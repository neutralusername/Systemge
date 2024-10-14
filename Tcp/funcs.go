package Tcp

import (
	"errors"
	"io"
	"net"
	"os"
	"syscall"
	"time"
)

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
