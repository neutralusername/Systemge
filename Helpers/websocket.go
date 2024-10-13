package Helpers

import (
	"errors"
	"io"
	"net"

	"github.com/gorilla/websocket"
)

func IsWebsocketConnClosedErr(err error) bool {
	if err == nil {
		return false
	}

	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		return true
	}

	if websocket.IsUnexpectedCloseError(err) {
		return true
	}

	if errors.Is(err, websocket.ErrCloseSent) {
		return true
	}

	if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
		return true
	}

	if nErr, ok := err.(net.Error); ok && !nErr.Timeout() {
		return true
	}

	return false
}
