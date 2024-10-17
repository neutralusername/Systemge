package SingleRequestServer

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/connectionTcp"
	"github.com/neutralusername/systemge/connectionWebsocket"
	"github.com/neutralusername/systemge/systemge"
)

func AsyncMessage[B any](connection systemge.Connection[B], data B, sendTimeoutNs int64) error {
	return connection.Write(data, sendTimeoutNs)
}

func AsyncMessageTcp(tcpClientConfig *configs.TcpClient, tcpBufferedReaderConfig *configs.TcpBufferedReader, data []byte, sendTimeoutNs int64) error {
	connection, err := connectionTcp.EstablishConnection(tcpBufferedReaderConfig, tcpClientConfig)
	if err != nil {
		return err
	}
	defer connection.Close()
	return connection.Write(data, sendTimeoutNs)
}

func AsyncMessageWebsocket(websocketClientConfig *configs.TcpClient, handshakeTimeoutNs int64, data []byte, sendTimeoutNs int64) error {
	connection, err := connectionWebsocket.EstablishConnection(websocketClientConfig, handshakeTimeoutNs)
	if err != nil {
		return err
	}
	defer connection.Close()
	return connection.Write(data, sendTimeoutNs)
}

func SyncRequest[B any](connection systemge.Connection[B], data B, sendTimeoutNs, readTimeoutNs int64) (B, error) {
	err := connection.Write(data, sendTimeoutNs)
	if err != nil {
		var nilValue B
		return nilValue, err
	}
	response, err := connection.Read(readTimeoutNs)
	if err != nil {
		var nilValue B
		return nilValue, err
	}
	return response, nil
}

func SyncRequestTcp(tcpClientConfig *configs.TcpClient, tcpBufferedReaderConfig *configs.TcpBufferedReader, data []byte, sendTimeoutNs, readTimeoutNs int64) ([]byte, error) {
	connection, err := connectionTcp.EstablishConnection(tcpBufferedReaderConfig, tcpClientConfig)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	err = connection.Write(data, sendTimeoutNs)
	if err != nil {
		return nil, err
	}
	response, err := connection.Read(readTimeoutNs)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func SyncRequestWebsocket(websocketClientConfig *configs.TcpClient, handshakeTimeoutNs int64, data []byte, sendTimeoutNs, readTimeoutNs int64) ([]byte, error) {
	connection, err := connectionWebsocket.EstablishConnection(websocketClientConfig, handshakeTimeoutNs)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	err = connection.Write(data, sendTimeoutNs)
	if err != nil {
		return nil, err
	}
	response, err := connection.Read(readTimeoutNs)
	if err != nil {
		return nil, err
	}
	return response, nil
}
