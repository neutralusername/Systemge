package listenerWebsocket

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/connectionWebsocket"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/systemge"
)

// wip
func Connect(
	tcpClientConfig *configs.TcpClient,
	incomingDataByteLimit uint64,
	timeoutNs int64,
) (systemge.Connection[[]byte], error) {

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = time.Duration(timeoutNs)
	netDialer := &net.Dialer{
		Timeout: time.Duration(timeoutNs),
	}
	dialer.NetDialContext = netDialer.DialContext

	if tcpClientConfig.TlsCert != "" {
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(tcpClientConfig.TlsCert))
		if !ok {
			return nil, fmt.Errorf("failed to parse TLS certificate")
		}

		dialer.TLSClientConfig = &tls.Config{
			RootCAs: roots,
		}

		if tcpClientConfig.Domain != "" {
			dialer.TLSClientConfig.ServerName = tcpClientConfig.Domain
		}
	}

	scheme := "ws"
	if tcpClientConfig.TlsCert != "" {
		scheme = "wss"
	}

	url := fmt.Sprintf("%s://%s", scheme, tcpClientConfig.Domain+":"+helpers.Uint16ToString(tcpClientConfig.Port))

	headers := http.Header{}

	if tcpClientConfig.Domain != "" {
		headers.Set("Host", tcpClientConfig.Domain)
	}

	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		return nil, err
	}

	return connectionWebsocket.New(conn, incomingDataByteLimit)
}
