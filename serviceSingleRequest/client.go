package serviceSingleRequest

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/connectionChannel"
	"github.com/neutralusername/systemge/connectionTcp"
	"github.com/neutralusername/systemge/connectionWebsocket"
	"github.com/neutralusername/systemge/systemge"
)

func AsyncMessage[D any](connection systemge.Connection[D], data D, sendTimeoutNs int64) error {
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

func AsyncMessageWebsocket(tcpClientConfig *configs.TcpClient, handshakeTimeoutNs int64, data []byte, sendTimeoutNs int64) error {
	connection, err := connectionWebsocket.EstablishConnection(tcpClientConfig, handshakeTimeoutNs)
	if err != nil {
		return err
	}
	defer connection.Close()
	return connection.Write(data, sendTimeoutNs)
}

func AsyncMessageChanne[D any](channelListenerConnectionReuqest chan<- *connectionChannel.ConnectionRequest[D], data D, sendTimeoutNs int64) error {
	connection, err := connectionChannel.EstablishConnection(channelListenerConnectionReuqest, sendTimeoutNs)
	if err != nil {
		return err
	}
	defer connection.Close()
	return connection.Write(data, sendTimeoutNs)
}

func SyncRequest[D any](connection systemge.Connection[D], data D, sendTimeoutNs, readTimeoutNs int64) (D, error) {
	err := connection.Write(data, sendTimeoutNs)
	if err != nil {
		var nilValue D
		return nilValue, err
	}
	response, err := connection.Read(readTimeoutNs)
	if err != nil {
		var nilValue D
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

func SyncRequestWebsocket(tcpClientConfig *configs.TcpClient, handshakeTimeoutNs int64, data []byte, sendTimeoutNs, readTimeoutNs int64) ([]byte, error) {
	connection, err := connectionWebsocket.EstablishConnection(tcpClientConfig, handshakeTimeoutNs)
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

func SyncRequestChanne[D any](channelListenerConnectionReuqest chan<- *connectionChannel.ConnectionRequest[D], data D, sendTimeoutNs, readTimeoutNs int64) (D, error) {
	connection, err := connectionChannel.EstablishConnection(channelListenerConnectionReuqest, sendTimeoutNs)
	if err != nil {
		var nilValue D
		return nilValue, err
	}
	defer connection.Close()
	err = connection.Write(data, sendTimeoutNs)
	if err != nil {
		var nilValue D
		return nilValue, err
	}
	response, err := connection.Read(readTimeoutNs)
	if err != nil {
		var nilValue D
		return nilValue, err
	}
	return response, nil
}
