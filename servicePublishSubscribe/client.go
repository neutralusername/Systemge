package servicePublishSubscribe

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/connectionChannel"
	"github.com/neutralusername/systemge/connectionTcp"
	"github.com/neutralusername/systemge/connectionWebsocket"
	"github.com/neutralusername/systemge/systemge"
)

func NewPublishSubscribeClientTcp(
	tcpClientConfig *configs.TcpClient,
	tcpBufferedReaderConfig *configs.TcpBufferedReader,
	sendTimeoutNs int64,
) (systemge.Connection[[]byte], error) {

	connection, err := connectionTcp.EstablishConnection(tcpBufferedReaderConfig, tcpClientConfig)
	if err != nil {
		return err
	}

}

func NewPublishSubscribeClientWebsocket(
	tcpClientConfig *configs.TcpClient,
	handshakeTimeoutNs int64,
	incomingDataByteLimit uint64,
	sendTimeoutNs int64,
) error {

	connection, err := connectionWebsocket.EstablishConnection(tcpClientConfig, handshakeTimeoutNs, incomingDataByteLimit)
	if err != nil {
		return err
	}
	defer connection.Close()

}

func AsyncMessageChanne[D any](
	channelListenerConnectionReuqest chan<- *connectionChannel.ConnectionRequest[D],
	sendTimeoutNs int64,
) error {

	connection, err := connectionChannel.EstablishConnection(channelListenerConnectionReuqest, sendTimeoutNs)
	if err != nil {
		return err
	}
	defer connection.Close()

}
