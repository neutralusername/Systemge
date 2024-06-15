package Client

import (
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
)

func (client *Client) tcpExchange(message *Message.Message, address string) (*Message.Message, error) {
	netConn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, Utilities.NewError("Error connecting to server", err)
	}
	err = Utilities.TcpSend(netConn, message.Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Utilities.NewError("Error sending message", err)
	}
	responseBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Utilities.NewError("Error receiving response", err)
	}
	message = Message.Deserialize(responseBytes)
	if message == nil || message.GetTopic() != "resolution" {
		netConn.Close()
		return nil, Utilities.NewError("Invalid response \""+string(responseBytes)+"\"", nil)
	}
	return message, nil
}
