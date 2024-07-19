package Tcp

import (
	"Systemge/Error"
	"Systemge/Message"
	"net"
	"time"
)

const ENDOFMESSAGE = "\x04"

func Exchange(netConn net.Conn, message *Message.Message, timeoutMs uint64, byteLimit uint64) (*Message.Message, error) {
	err := Send(netConn, message.Serialize(), timeoutMs)
	if err != nil {
		return nil, Error.New("Error sending message", err)
	}
	responseBytes, err := Receive(netConn, timeoutMs, byteLimit)
	if err != nil {
		return nil, Error.New("Error receiving response", err)
	}
	responseMessage := Message.Deserialize(responseBytes)
	if responseMessage == nil {
		return nil, Error.New("Error deserializing response", nil)
	}
	return responseMessage, nil
}

func Send(netConn net.Conn, bytes []byte, timeoutMs uint64) error {
	if netConn == nil {
		return Error.New("net.Conn is nil", nil)
	}
	if timeoutMs > 0 {
		netConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	} else {
		netConn.SetWriteDeadline(time.Time{})
	}
	_, err := netConn.Write(append(bytes, []byte(ENDOFMESSAGE)...))
	if err != nil {
		return err
	}
	return nil
}

func Receive(netConn net.Conn, timeoutMs uint64, bytelimit uint64) ([]byte, error) {
	if netConn == nil {
		return nil, Error.New("net.Conn is nil", nil)
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
		if err != nil {
			return nil, err
		}
		if buffer[n-1] == []byte(ENDOFMESSAGE)[0] {
			break
		} else {
			len += uint64(n)
			if bytelimit > 0 && len > bytelimit {
				return nil, Error.New("Message exceeds byte limit", nil)
			}
			message = append(message, buffer[:n]...)
		}
	}
	return message, nil
}
