package servicePublishSubscribe

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/connectionChannel"
	"github.com/neutralusername/systemge/connectionTcp"
	"github.com/neutralusername/systemge/connectionWebsocket"
	"github.com/neutralusername/systemge/systemge"
)

type Client[D any] struct {
	connection systemge.Connection[D]
}

func NewPublishSubscribeClientTcp(
	tcpClientConfig *configs.TcpClient,
	tcpBufferedReaderConfig *configs.TcpBufferedReader,
	sendTimeoutNs int64,
) (*Client[[]byte], error) {

	connection, err := connectionTcp.EstablishConnection(tcpBufferedReaderConfig, tcpClientConfig)
	if err != nil {
		return nil, err
	}
	client := &Client[[]byte]{
		connection: connection,
	}

	return client, nil
}

func NewPublishSubscribeClientWebsocket(
	tcpClientConfig *configs.TcpClient,
	handshakeTimeoutNs int64,
	incomingDataByteLimit uint64,
	sendTimeoutNs int64,
) (*Client[[]byte], error) {

	connection, err := connectionWebsocket.EstablishConnection(tcpClientConfig, handshakeTimeoutNs, incomingDataByteLimit)
	if err != nil {
		return nil, err
	}
	client := &Client[[]byte]{
		connection: connection,
	}

	return client, nil
}

func NewPublishSubscribeClientChannel[D any](
	channelListenerConnectionReuqest chan<- *connectionChannel.ConnectionRequest[D],
	sendTimeoutNs int64,
) (*Client[D], error) {

	connection, err := connectionChannel.EstablishConnection(channelListenerConnectionReuqest, sendTimeoutNs)
	if err != nil {
		return nil, err
	}
	client := &Client[D]{
		connection: connection,
	}

	return client, nil
}
