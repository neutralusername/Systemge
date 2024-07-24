package Tcp

import (
	"Systemge/Error"
	"Systemge/Message"
	"net"
	"time"
)

const ENDOFMESSAGE = "\x04"

func Exchange(netConn net.Conn, messageBytes []byte, timeoutMs uint64, byteLimit uint64) (*Message.Message, uint64, uint64, error) {
	bytesSent, err := Send(netConn, messageBytes, timeoutMs)
	if err != nil {
		return nil, 0, 0, Error.New("Error sending message", err)
	}
	responseBytes, bytesReceived, err := Receive(netConn, timeoutMs, byteLimit)
	if err != nil {
		return nil, 0, 0, Error.New("Error receiving response", err)
	}
	responseMessage := Message.Deserialize(responseBytes)
	if responseMessage == nil {
		return nil, 0, 0, Error.New("Error deserializing response", nil)
	}
	return responseMessage, bytesSent, bytesReceived, nil
}

func Send(netConn net.Conn, bytes []byte, timeoutMs uint64) (uint64, error) {
	if netConn == nil {
		return 0, Error.New("net.Conn is nil", nil)
	}
	if timeoutMs > 0 {
		netConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	} else {
		netConn.SetWriteDeadline(time.Time{})
	}
	bytesSend, err := netConn.Write(append(bytes, []byte(ENDOFMESSAGE)...))
	if err != nil {
		return 0, err
	}
	return uint64(bytesSend), nil
}

func Receive(netConn net.Conn, timeoutMs uint64, bytelimit uint64) ([]byte, uint64, error) {
	if netConn == nil {
		return nil, 0, Error.New("net.Conn is nil", nil)
	}
	buffer := make([]byte, 1)
	message := make([]byte, 0)
	var len uint64 = 0
	if timeoutMs > 0 {
		netConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	} else {
		netConn.SetReadDeadline(time.Time{})
	}
	for {
		n, err := netConn.Read(buffer)
		len += uint64(n)
		if err != nil {
			return nil, 0, err
		}
		if buffer[n-1] == []byte(ENDOFMESSAGE)[0] {
			break
		} else {
			if bytelimit > 0 && len > bytelimit {
				return nil, 0, Error.New("Message exceeds byte limit", nil)
			}
			message = append(message, buffer[:n]...)
		}
	}
	return message, len, nil
}
