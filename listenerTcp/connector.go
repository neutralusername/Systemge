package listenerTcp

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/systemge"
)

type connector struct {
	tcpBufferedReaderConfig *configs.TcpBufferedReader
	tcpClientConfig         *configs.TcpClient
}

func NewConnector(
	tcpBufferedReaderConfig *configs.TcpBufferedReader,
	tcpClientConfig *configs.TcpClient,
) systemge.Connector[[]byte] {
	return &connector{
		tcpBufferedReaderConfig: tcpBufferedReaderConfig,
		tcpClientConfig:         tcpClientConfig,
	}
}

func (connector *connector) Connect(timeoutNs int64) (systemge.Connection[[]byte], error) {
	return Connect(connector.tcpBufferedReaderConfig, connector.tcpClientConfig, timeoutNs)
}
