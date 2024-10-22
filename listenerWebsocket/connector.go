package listenerWebsocket

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/systemge"
)

type connector struct {
	tcpClientConfig       *configs.TcpClient
	incomingDataByteLimit uint64
}

func NewConnector(
	tcpClientConfig *configs.TcpClient,
	incomingDataByteLimit uint64,
) systemge.Connector[[]byte, systemge.Connection[[]byte]] {
	return &connector{
		tcpClientConfig:       tcpClientConfig,
		incomingDataByteLimit: incomingDataByteLimit,
	}
}

func (connector *connector) Connect(timeoutNs int64) (systemge.Connection[[]byte], error) {
	return Connect(connector.tcpClientConfig, connector.incomingDataByteLimit, timeoutNs)
}
