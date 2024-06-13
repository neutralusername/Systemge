package TCP

import (
	"Systemge/Error"
	"net"
	"time"
)

func Send(netConn net.Conn, msg []byte, timeoutMs int) error {
	if netConn == nil {
		return Error.New("net.Conn is nil", nil)
	}
	if timeoutMs > 0 {
		netConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	} else {
		netConn.SetWriteDeadline(time.Time{})
	}
	_, err := netConn.Write(append(msg, []byte(ENDOFMESSAGE)...))
	if err != nil {
		return err
	}
	return nil
}

func Receive(netConn net.Conn, timeoutMs int) ([]byte, error) {
	if netConn == nil {
		return nil, Error.New("net.Conn is nil", nil)
	}
	buffer := make([]byte, 1)
	message := make([]byte, 0)
	if timeoutMs > 0 {
		netConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	} else {
		netConn.SetReadDeadline(time.Time{})
	}
	for {
		n, err := netConn.Read(buffer)
		if err != nil {
			return nil, err
		}
		if buffer[n-1] == []byte(ENDOFMESSAGE)[0] {
			break
		} else {
			message = append(message, buffer[:n]...)
		}
	}
	return message, nil
}
