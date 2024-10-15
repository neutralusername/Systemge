package Helpers

import (
	"errors"
	"io"
	"net"
	"os"
	"syscall"
)

func IsNetConnClosedErr(err error) bool {
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
